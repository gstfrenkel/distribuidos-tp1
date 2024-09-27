package review

import (
	"fmt"
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
	fmt.Println("Holaaaa")

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

	fmt.Println("Chauuuu")
}
