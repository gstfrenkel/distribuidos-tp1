package indie

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const (
	query2 uint8 = iota
	query3
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
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, nil, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
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
	genre := getGenre(f)
	gameReleases := msg.ToGameReleasesMessage(genre)
	b, err := gameReleases.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	headers := map[string]any{amqp.MessageIdHeader: message.GameReleaseID}
	if err = f.w.Broker.Publish(f.w.Outputs[query2].Exchange, f.w.Outputs[query3].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	gameNames := msg.ToGameNamesMessage(genre)
	headers[amqp.MessageIdHeader] = message.GameNameID
	for _, game := range gameNames {
		b, err = game.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := worker.ShardGameId(game.GameId, f.w.Outputs[query3].Key, f.w.Outputs[query3].Consumers)
		if err = f.w.Broker.Publish(f.w.Outputs[query3].Exchange, k, b, headers); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}

func getGenre(f *filter) string {
	return f.w.Query.(string)
}
