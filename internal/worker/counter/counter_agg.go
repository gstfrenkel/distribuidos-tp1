package counter

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type filter struct {
	w     *worker.Worker
	games message.GameNames
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w, games: nil}, nil
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
		f.publish()
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}
		f.games = append(f.games, msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish() {
	if f.games == nil {
		return
	}
	logs.Logger.Infof("Query 4 results: %v", f.games)
	b, err := f.games.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	headers := map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID), amqp.OriginIdHeader: amqp.Query4originId}
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
