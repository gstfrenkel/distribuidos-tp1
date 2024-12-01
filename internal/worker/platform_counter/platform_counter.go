package platform_counter

import (
	"strings"

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

func New() (worker.Filter, error) {
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

	if _, exists := f.counters[headers.ClientId]; !exists {
		f.counters[headers.ClientId] = &message.Platform{Windows: 0, Linux: 0, Mac: 0}
	}
	clientCounter := f.counters[headers.ClientId]

	headers = headers.WithOriginId(amqp.Query1originId) // TODO: Add sequence ID

	switch headers.MessageId {
	case message.EofMsg:
		f.publish(headers)
		delete(f.counters, headers.ClientId)

		if !f.agg {
			if _, err := f.w.HandleEofMessage(delivery.Body, headers); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
			}
		}
	case message.PlatformID:
		msg, err := message.PlatfromFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			clientCounter.Increment(msg)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(headers amqp.Header) {
	platforms := f.counters[headers.ClientId]

	if platforms.IsEmpty() {
		return
	}

	b, err := platforms.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	output := f.w.Outputs[0]
	if f.agg {
		output, err = shard.AggregatorOutput(output, headers.ClientId)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		}
	}

	logs.Logger.Infof("Output: %v", output)

	headers = headers.WithMessageId(message.PlatformID)
	if err = f.w.Broker.Publish(output.Exchange, output.Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
