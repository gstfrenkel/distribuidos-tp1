package indie

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
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

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var err error

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds, err = f.w.HandleEofMessage(delivery.Body, headers.WithOriginId(amqp.GameOriginId))
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	case message.GameIdMsg:
		msg, err := message.GameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			f.publish(headers, msg)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(headers amqp.Header, msg message.Game) {
	genre := getGenre(f)
	gameReleases := msg.ToGameReleasesMessage(genre)
	b, err := gameReleases.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	output := f.w.Outputs[query2]
	key := worker.ShardSequenceId(headers.SequenceId, output.Key, output.Consumers)

	if err = f.w.Broker.Publish(output.Exchange, key, b, headers.WithMessageId(message.GameReleaseID)); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	gameNames := msg.ToGameNamesMessage(genre)
	for _, game := range gameNames {
		b, err = game.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		output = f.w.Outputs[query3]
		key = worker.ShardGameId(game.GameId, output.Key, output.Consumers)
		if err = f.w.Broker.Publish(output.Exchange, key, b, headers.WithMessageId(message.GameNameID)); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}

func getGenre(f *filter) string {
	return f.w.Query.(string)
}
