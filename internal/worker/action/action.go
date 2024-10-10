package action

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

var (
	headersEof = map[string]any{amqp.OriginIdHeader: amqp.GameOriginId}
	headers    = map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID)}
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

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, headersEof, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	} else if messageId == message.GameIdMsg {
		msg, err := message.GameFromBytes(delivery.Body)
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
			k := worker.ShardGameId(game.GameId, output.Key, output.Consumers)
			if err = f.w.Broker.Publish(output.Exchange, k, b, headers); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
			}
		}
	}
}
