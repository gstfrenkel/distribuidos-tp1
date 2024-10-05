package indie

import (
	"fmt"
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
			fmt.Printf("\n%s\n", errors.FailedToPublish.Error())
		}
	} else if messageId == message.GameIdMsg {
		msg, err := message.GameFromBytes(reviewDelivery.Body)
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg)
	} else {
		fmt.Printf(errors.InvalidMessageId.Error(), messageId)
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

	if err = f.w.Broker.Publish(routes[playtimeQueue].Exchange, routes[playtimeQueue].Key, uint8(message.GameReleaseID), b); err != nil {
		fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	gameNames := msg.ToGameNamesMessage(genre)
	for _, game := range gameNames {
		b, err = game.ToBytes()
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := routes[positiveQueue].Key
		if routes[positiveQueue].Consumers > 0 {
			k = worker.ShardGameId(game.GameId, k, routes[positiveQueue].Consumers)
		}
		if err = f.broker.Publish(routes[positiveQueue].Exchange, k, uint8(message.GameNameID), b); err != nil {
			fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}

func getGenre(f *filter) string {
	return f.w.Query.(string)
}
