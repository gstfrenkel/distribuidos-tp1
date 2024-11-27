package platform_counter

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

type filter struct {
	counters map[string]*message.Platform
	w        *worker.Worker
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
		logs.Logger.Infof("Received EOF for client %s! Sending platform count: %v", headers.ClientId, clientCounter)

		f.publish(headers)
		delete(f.counters, headers.ClientId)

		_, err := f.w.HandleEofMessage(delivery.Body, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
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

	logs.Logger.Infof("Sending %v for client %s with key %s", *platforms, headers.ClientId, f.w.Outputs[0].Key)

	headers = headers.WithMessageId(message.PlatformID)
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
