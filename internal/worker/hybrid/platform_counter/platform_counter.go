package platform_counter

import (
	"strings"
	"tp1/pkg/recovery"

	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

type filter struct {
	counters map[string]*message.Platform
	w        *worker.Worker
	agg      bool
}

func New() (worker.W, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{
		w:        w,
		counters: make(map[string]*message.Platform),
	}, nil
}

func (f *filter) Init() error {
	return f.w.Init()
}

func (f *filter) Start() {
	f.agg = strings.Contains(f.w.Outputs[0].Key, "%d")

	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = f.processEof(delivery.Body, headers, false)
	case message.PlatformID:
		f.processPlatform(delivery.Body, headers.ClientId)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) processPlatform(msgBytes []byte, clientId string) {
	if _, exists := f.counters[clientId]; !exists {
		f.counters[clientId] = &message.Platform{Windows: 0, Linux: 0, Mac: 0}
	}

	msg, err := message.PlatfromFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else {
		f.counters[clientId].Increment(msg)
	}
}

func (f *filter) processEof(msgBytes []byte, headers amqp.Header, recovery bool) []sequence.Destination {
	headers = headers.WithOriginId(amqp.Query1originId)
	var sequenceIds []sequence.Destination
	if !recovery {
		sequenceIds = f.publish(headers)
	}
	delete(f.counters, headers.ClientId)

	if !f.agg {
		eofSqIds, err := f.w.HandleEofMessage(msgBytes, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		} else {
			sequenceIds = append(sequenceIds, eofSqIds...)
		}
	}

	return sequenceIds
}

func (f *filter) publish(headers amqp.Header) []sequence.Destination {
	var sequenceIds []sequence.Destination
	platforms := f.counters[headers.ClientId]

	if platforms.IsEmpty() {
		return sequenceIds
	}

	b, err := platforms.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return sequenceIds
	}

	output := f.w.Outputs[0]
	if f.agg {
		output, err = shard.AggregatorOutput(output, headers.ClientId)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		}
	}

	sequenceId := f.w.NextSequenceId(output.Key)
	headers = headers.WithMessageId(message.PlatformID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))
	sequenceIds = append(sequenceIds, sequence.DstNew(output.Key, sequenceId))

	if err = f.w.Broker.Publish(output.Exchange, output.Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return sequenceIds
}

func (f *filter) recover() {
	ch := make(chan recovery.Message, worker.ChanSize)
	go f.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			f.processEof(recoveredMsg.Message(), recoveredMsg.Header(), true)
		case message.PlatformID:
			f.processPlatform(recoveredMsg.Message(), recoveredMsg.Header().ClientId)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
