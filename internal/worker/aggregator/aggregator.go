package aggregator

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

type processor interface {
	sendEof(headers amqp.Header) []sequence.Destination
	reset(clientId string)
	publish(headers amqp.Header) []sequence.Destination
	save(msgBytes []byte, clientId string)
}

type aggregator struct {
	w         *worker.Worker
	batchSize uint16
	eofsRecv  map[string]uint8 // <clientId, eofsReceived>
}

// newAggregator creates a new aggregator.
// Field batchSize is not set here. It should be set by the caller.
func newAggregator() (*aggregator, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &aggregator{
		w:        w,
		eofsRecv: make(map[string]uint8),
	}, nil
}

func (a *aggregator) processEof(instance processor, headers amqp.Header, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	a.eofsRecv[headers.ClientId]++
	if a.eofsReached(headers) {
		if !recovery {
			sequenceIds = instance.publish(headers)
		}
		sequenceIds = append(sequenceIds, instance.sendEof(headers)...)
		a.reset(headers.ClientId)
		instance.reset(headers.ClientId)
	}
	return sequenceIds
}

func (a *aggregator) reset(clientId string) {
	delete(a.eofsRecv, clientId)
}

func (a *aggregator) recover(instance processor, msgId message.ID) {
	ch := make(chan recovery.Message, worker.ChanSize)
	go a.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			a.processEof(instance, recoveredMsg.Header(), true)
		case msgId:
			instance.save(recoveredMsg.Message(), recoveredMsg.Header().ClientId)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}

func (a *aggregator) eofsReached(headers amqp.Header) bool {
	return a.eofsRecv[headers.ClientId] >= a.w.ExpectedEofs
}

func shardOutput(output amqp.Destination, clientId string) amqp.Destination {
	output, err := shard.AggregatorOutput(output, clientId)

	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}
	return output
}
