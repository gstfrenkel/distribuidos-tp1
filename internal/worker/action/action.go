package action

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
	games := msg.ToGameNamesMessage(f.w.Query.(string))
	for _, game := range games {
		b, err := game.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		for _, output := range f.w.Outputs {
			k := output.Key
			if output.Consumers > 0 {
				k = worker.ShardGameId(game.GameId, k, output.Consumers)
			}
			if err = f.w.Broker.Publish(output.Exchange, k, uint8(message.GameNameID), b); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
			}
		}
	}
}
