package review

import (
	"fmt"
	"time"
	"tp1/internal/errors"
	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/message"
	"tp1/pkg/message/review"
	"tp1/pkg/message/scored_review"
)

var (
	exchange          = "reviews"
	positiveConsumers = 1
	positiveKey       = "p%d"
	negativeConsumers = 1
	negativeKey       = "n%d"
)

type Filter struct {
	config config.Config
	broker broker.MessageBroker
}

func New() (*Filter, error) {
	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	b, err := amqpconn.New()
	if err != nil {
		return nil, err
	}

	return &Filter{
		config: cfg,
		broker: b,
	}, nil
}

func (f Filter) Init() error {
	if err := f.broker.ExchangeDeclare(f.config.String("exchange.name", "reviews"), f.config.String("exchange.kind", "direct")); err != nil {
		return err
	} else if _, err = f.broker.QueuesDeclare(f.queues()...); err != nil {
		return err
	} else if err = f.broker.QueuesBind(f.binds()...); err != nil {
		return err
	}
	return nil
}

func (f Filter) Start() {
	defer f.broker.Close()

	b, _ := review.Message{
		{GameId: 1, GameName: "Game1", Text: "Great game", Score: 1},
		{GameId: 1, GameName: "Game1", Text: "Great game x2", Score: 1},
		{GameId: 1, GameName: "Game1", Text: "Bad game", Score: -1},
		{GameId: 2, GameName: "Game2", Text: "Bad game", Score: -1},
	}.ToBytes()

	time.Sleep(time.Second * 10)

	_ = f.broker.Publish("gateway", "reviews", uint8(message.ReviewIdMsg), b)

	exchange = f.config.String("exchange.name", "reviews")
	positiveKey = f.config.String("positive-reviews-sh.key", "p%d")
	negativeKey = f.config.String("negative-reviews-sh.key", "n%d")
	positiveConsumers = f.config.Int("positive-reviews-sh.consumers", 1)
	negativeConsumers = f.config.Int("negative-reviews-sh.consumers", 1)

	reviewChan, err := f.broker.Consume(f.config.String("gateway.queue-name", "reviews"), "", true, false)
	if err != nil {
		println(err.Error())
		return
	}

	for reviewDelivery := range reviewChan {
		messageId := message.ID(reviewDelivery.Headers[amqpconn.MessageIdHeader].(uint8))

		if messageId != message.EofMsg && messageId != message.ReviewIdMsg {
			fmt.Printf(errors.InvalidMessageId.Error(), messageId)
			continue
		}

		if messageId == message.EofMsg {
			if err = f.broker.Publish(exchange, "", uint8(message.EofMsg), reviewDelivery.Body); err != nil {
				fmt.Printf(errors.FailedToPublish.Error())
				return
			}
			continue
		}

		msg, err := review.FromBytes(reviewDelivery.Body)
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		f.publish(msg)
	}

	fmt.Println("Chauuuu")
}

func (f Filter) publish(msg review.Message) {
	b, err := msg.ToPositiveReviewWithTextMessage().ToBytes()
	if err != nil {
		fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else if err = f.broker.Publish(exchange, "", uint8(message.EofMsg), b); err != nil {
		fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	f.shardPublish(msg.ToPositiveReviewMessage(), positiveKey, positiveConsumers)
	f.shardPublish(msg.ToNegativeReviewMessage(), negativeKey, negativeConsumers)
}

func (f Filter) shardPublish(reviews scored_review.Messages, k string, consumers int) {
	for _, rv := range reviews {
		b, err := rv.ToBytes()
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		key := fmt.Sprintf(k, rv.GameId%int64(consumers))
		if err = f.broker.Publish(exchange, key, uint8(message.EofMsg), b); err != nil {
			fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}
