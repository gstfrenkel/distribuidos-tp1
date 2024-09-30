package gateway

import (
	"net"

	"tp1/internal/gateway/rabbit"
	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"

	"github.com/rabbitmq/amqp091-go"
)

type Gateway struct {
	Config       config.Config
	broker       broker.MessageBroker
	reviewsQueue amqp091.Queue
	gamesQueue   amqp091.Queue
	exchange     string
	Listener     net.Listener
}

func New() (*Gateway, error) {
	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	b, err := amqpconn.NewBroker()
	if err != nil {
		return nil, err
	}

	reviewsQ, gamesQ, err := rabbit.CreateQueues(err, b, cfg)
	if err != nil {
		return nil, err
	}

	exchangeName, err := rabbit.CreateExchange(cfg, err, b)
	if err != nil {
		return nil, err
	}

	err = rabbit.BindQueuesToExchange(err, b, reviewsQ.Name, gamesQ.Name, cfg, exchangeName)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		Config:       cfg,
		broker:       b,
		reviewsQueue: reviewsQ,
		gamesQueue:   gamesQ,
		exchange:     exchangeName,
	}, nil
}

func (g Gateway) Start() {
	defer g.broker.Close()

	err := CreateGatewaySocket(&g)
	if err != nil { //TODO handle
		return
	}

	defer g.Listener.Close() //TODO handle

	err = ListenForNewClients(&g)
	if err != nil {
		return
	}
}

func (g Gateway) End() {
	g.broker.Close()
}
