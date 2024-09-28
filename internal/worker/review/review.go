package review

import (
	"fmt"
	msg "tp1/pkg/message"
	"tp1/pkg/message/scored_review"
	"tp1/pkg/message/utils"

	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
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

	original := scored_review.New(map[utils.Key]int64{
		utils.Key{GameId: 1, GameName: "Game1"}: 4,
		utils.Key{GameId: 2, GameName: "Game1"}: 1,
		utils.Key{GameId: 3, GameName: "Game1"}: 8,
	})

	b, err := original.ToBytes()
	if err != nil {
		fmt.Printf("\n\nErrorrrrrr: %s\n\n", err.Error())
		return
	}
	if err = f.broker.Publish("reviews", "", uint8(msg.PositiveReviewID), b); err != nil {
		fmt.Printf("\n\nErrorrrrrr: %s\n\n", err.Error())
		return
	}

	fmt.Println("Chauuuu")
}
