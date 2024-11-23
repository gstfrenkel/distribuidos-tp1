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

func (f *filter) Process(delivery amqp.Delivery, _ amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)

	if _, exists := f.counters[clientId]; !exists {
		f.counters[clientId] = &message.Platform{Windows: 0, Linux: 0, Mac: 0}
	}
	clientCounter := f.counters[clientId]

	if messageId == message.EofMsg {
		logs.Logger.Infof("Received EOF for client %s! Sending platform count: %v", clientId, clientCounter)

		f.publish(clientId)
		delete(f.counters, clientId)

		headers := map[string]any{amqp.ClientIdHeader: clientId}
		_, err := f.w.HandleEofMessage(delivery.Body, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}

	} else if messageId == message.PlatformID {
		msg, err := message.PlatfromFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			clientCounter.Increment(msg)
		}
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(clientId string) {
	platforms := f.counters[clientId]

	if platforms.IsEmpty() {
		return
	}

	b, err := platforms.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	headers := map[string]any{
		amqp.MessageIdHeader: uint8(message.PlatformID),
		amqp.OriginIdHeader:  amqp.Query1originId,
		amqp.ClientIdHeader:  clientId,
	}
	logs.Logger.Infof("Sending %v for client %s with key %s", *platforms, clientId, f.w.Outputs[0].Key)

	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
