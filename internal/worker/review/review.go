package review

import (
	"fmt"

	"tp1/pkg/amqp"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
)

type Filter struct {
	config config.Config
	broker *amqp.MessageBroker
}

func New() (*Filter, error) {
	fmt.Println("Holaaaa")

	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	broker, err := amqp.New()
	if err != nil {
		return nil, err
	}

	return &Filter{
		config: cfg,
		broker: broker,
	}, nil
}

func (f Filter) Start() {
	defer f.broker.Close()

	fmt.Println("Chauuuu")
}
