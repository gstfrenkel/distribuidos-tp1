package rabbit

import (
	"github.com/rabbitmq/amqp091-go"
	"tp1/pkg/broker"
	"tp1/pkg/config"
)

func CreateQueues(err error, b broker.MessageBroker, cfg config.Config) (amqp091.Queue, amqp091.Queue, error) {
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
	return reviewsQ[0], gamesQ[0], nil
}

func CreateExchange(cfg config.Config, err error, b broker.MessageBroker) (string, error) {
	exchangeName := cfg.String("rabbitmq.exchange_name", "e")
	if err := b.ExchangeDeclare(map[string]string{exchangeName: cfg.String("rabbitmq.exchange_type", "direct")}); err != nil {
		b.Close()
		return "", err
	}
	return exchangeName, nil
}

func BindQueuesToExchange(err error, b broker.MessageBroker, reviewsQ string, gamesQ string, cfg config.Config, exchangeName string) error {
	err = b.QueueBind(broker.QueueBind{
		Name:     reviewsQ,
		Key:      cfg.String("rabbitmq.reviews_routing_key", "r"),
		Exchange: exchangeName,
	})
	if err != nil {
		b.Close()
		return err
	}

	err = b.QueueBind(broker.QueueBind{
		Name:     gamesQ,
		Key:      cfg.String("rabbitmq.games_routing_key", "g"),
		Exchange: exchangeName,
	})
	if err != nil {
		b.Close()
		return err
	}

	return nil
}
