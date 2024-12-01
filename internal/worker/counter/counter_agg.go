package counter

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

type filter struct {
	w         *worker.Worker
	games     map[string]message.GameNames // <clientId, []GameName>
	batchSize uint16
	eofsRecv  map[string]uint8 // <clientId, eofsReceived>
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{
		w:         w,
		games:     make(map[string]message.GameNames),
		eofsRecv:  make(map[string]uint8),
		batchSize: uint16(w.Query.(float64)),
	}, nil
}

func (f *filter) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = f.processEof(headers, false)
	case message.GameNameID:
		f.saveGame(delivery.Body, headers.ClientId)
		sequenceIds = f.publish(headers, false)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) processEof(headers amqp.Header, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	f.eofsRecv[headers.ClientId]++
	if f.eofsReached(headers) {
		if !recovery {
			sequenceIds = f.publish(headers, true)
		}
		sequenceIds = append(sequenceIds, f.sendEof(headers)...)
		f.reset(headers.ClientId)
	}
	return sequenceIds
}

func (f *filter) publish(headers amqp.Header, eof bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	if len(f.games[headers.ClientId]) > 0 && (f.batchSizeReached(headers) || eof) {
		b, err := f.games[headers.ClientId].ToBytes()
		logs.Logger.Infof("Publishing batch of %d games for client %s", len(f.games), headers.ClientId)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return sequenceIds
		}

		sequenceIds = f.sendBatch(headers, b)
		f.games[headers.ClientId] = f.games[headers.ClientId][:0]
	}

	return sequenceIds
}

func (f *filter) sendBatch(headers amqp.Header, b []byte) []sequence.Destination {
	output := shardOutput(f.w.Outputs[0], headers.ClientId)
	key := output.Key
	sequenceId := f.w.NextSequenceId(key)
	headers = headers.WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId)).WithOriginId(amqp.Query4originId)

	if err := f.w.Broker.Publish(output.Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func (f *filter) sendEof(headers amqp.Header) []sequence.Destination {
	output := shardOutput(f.w.Outputs[0], headers.ClientId)
	sequenceIds, err := f.w.HandleEofMessage(amqp.EmptyEof, headers.WithOriginId(amqp.Query4originId), amqp.DestinationEof(output))
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Debugf("Eof message sent for client %s", headers.ClientId)
	return sequenceIds
}

func (f *filter) reset(clientId string) {
	delete(f.eofsRecv, clientId)
	delete(f.games, clientId)
}

func (f *filter) saveGame(msgBytes []byte, clientId string) {
	msg, err := message.GameNameFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	if _, ok := f.games[clientId]; !ok {
		f.games[clientId] = make(message.GameNames, 0)
	}

	f.games[clientId] = append(f.games[clientId], msg)
}

func (f *filter) recover() {
	ch := make(chan recovery.Message, worker.ChanSize)
	go f.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			f.processEof(recoveredMsg.Header(), true)
		case message.GameNameID:
			f.saveGame(recoveredMsg.Message(), recoveredMsg.Header().ClientId)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}

func (f *filter) batchSizeReached(headers amqp.Header) bool {
	return len(f.games[headers.ClientId]) >= int(f.batchSize)
}

func (f *filter) eofsReached(headers amqp.Header) bool {
	return f.eofsRecv[headers.ClientId] >= f.w.ExpectedEofs
}

func shardOutput(output amqp.Destination, clientId string) amqp.Destination {
	output, err := shard.AggregatorOutput(output, clientId)

	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}
	return output
}
