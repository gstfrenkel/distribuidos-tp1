package gateway

import (
	"github.com/rabbitmq/amqp091-go"
	"time"
	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
)

type Gateway struct {
	config       config.Config
	broker       broker.MessageBroker
	reviewsQueue amqp091.Queue
	gamesQueue   amqp091.Queue
	exchange     string
}

func New() (*Gateway, error) {
	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	b, err := amqpconn.New()
	if err != nil {
		return nil, err
	}

	reviewsQ, gamesQ, err := createQueues(err, b, cfg)
	if err != nil {
		return nil, err
	}

	exchangeName, err := createExchange(cfg, err, b)
	if err != nil {
		return nil, err
	}

	err = bindQueuesToExchange(err, b, reviewsQ.Name, gamesQ.Name, cfg, exchangeName)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		config:       cfg,
		broker:       b,
		reviewsQueue: reviewsQ,
		gamesQueue:   gamesQ,
		exchange:     exchangeName,
	}, nil
}

func createExchange(cfg config.Config, err error, b broker.MessageBroker) (string, error) {
	exchangeName := cfg.String("rabbitmq.exchange_name", "e")
	err = b.ExchangeDeclare(exchangeName, cfg.String("rabbitmq.exchange_type", "direct"))
	if err != nil {
		b.Close()
		return "", err
	}
	return exchangeName, nil
}

func bindQueuesToExchange(err error, b broker.MessageBroker, reviewsQ string, gamesQ string, cfg config.Config, exchangeName string) error {
	err = b.QueueBind(reviewsQ, cfg.String("rabbitmq.reviews_routing_key", "r"), exchangeName)
	if err != nil {
		b.Close()
		return err
	}

	err = b.QueueBind(gamesQ, cfg.String("rabbitmq.games_routing_key", "g"), exchangeName)
	if err != nil {
		b.Close()
		return err
	}

	return nil
}

func createQueues(err error, b broker.MessageBroker, cfg config.Config) (amqp091.Queue, amqp091.Queue, error) {
	reviewsQ, err := b.QueueDeclare(cfg.String("rabbitmq.reviews_q", "r"))
	if err != nil {
		b.Close()
		return amqp091.Queue{}, amqp091.Queue{}, err
	}

	gamesQ, err := b.QueueDeclare(cfg.String("rabbitmq.games_q", "g"))
	if err != nil {
		b.Close()
		return amqp091.Queue{}, amqp091.Queue{}, err
	}
	return reviewsQ, gamesQ, nil
}

func (g Gateway) Start() {
	defer g.broker.Close()
	time.Sleep(15 * time.Second)
}

func (g Gateway) End() {
	g.broker.Close()
}
