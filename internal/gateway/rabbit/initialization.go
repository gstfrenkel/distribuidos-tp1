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
	return reviewsQ, gamesQ, nil
}

func CreateExchange(cfg config.Config, err error, b broker.MessageBroker) (string, error) {
	exchangeName := cfg.String("rabbitmq.exchange_name", "e")
	err = b.ExchangeDeclare(exchangeName, cfg.String("rabbitmq.exchange_type", "direct"))
	if err != nil {
		b.Close()
		return "", err
	}
	return exchangeName, nil
}

func BindQueuesToExchange(err error, b broker.MessageBroker, reviewsQ string, gamesQ string, cfg config.Config, exchangeName string) error {
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
