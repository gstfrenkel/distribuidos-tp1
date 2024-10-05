package platform

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type filter struct {
	w *worker.Worker
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}
	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(reviewDelivery amqpconn.Delivery) {
	messageId := message.ID(reviewDelivery.Headers[amqpconn.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, reviewDelivery.Body, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	} else if messageId == message.GameIdMsg {
		msg, err := message.GameFromBytes(reviewDelivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(msg message.Game) {
	platforms := msg.ToPlatformMessage()
	b, err := platforms.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, uint8(message.PlatformID), b); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
