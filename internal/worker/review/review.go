package review

import (
	"fmt"

	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"

	"github.com/rabbitmq/amqp091-go"
)

type Filter struct {
	config   config.Config
	broker   broker.MessageBroker
	srcQueue amqp091.Queue
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

func (f Filter) Start() {
	defer f.broker.Close()

	fmt.Println("Chauuuu")
}
