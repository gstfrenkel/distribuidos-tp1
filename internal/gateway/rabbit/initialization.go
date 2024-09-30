package rabbit

import (
	"github.com/rabbitmq/amqp091-go"
	"tp1/pkg/broker"
	"tp1/pkg/config"
)

func CreateGatewayQueues(b broker.MessageBroker, cfg config.Config) (amqp091.Queue, amqp091.Queue, error) {
	reviewsAndGamesQ, err := b.QueueDeclare(cfg.String("rabbitmq.reviews_q", "r"), cfg.String("rabbitmq.games_q", "g"))
	if err != nil {
		b.Close()
		return amqp091.Queue{}, amqp091.Queue{}, err
	}

	return reviewsAndGamesQ[0], reviewsAndGamesQ[1], nil
}

func CreateGatewayExchange(cfg config.Config, b broker.MessageBroker) (string, error) {
	exchangeName := cfg.String("rabbitmq.exchange_name", "e")
	err := b.ExchangeDeclare(broker.Exchange{Name: exchangeName, Kind: cfg.String("rabbitmq.exchange_type", "direct")})
	if err != nil {
		b.Close()
		return "", err
	}
	return exchangeName, nil
}

func BindGatewayQueuesToExchange(b broker.MessageBroker, reviewsQ string, gamesQ string, cfg config.Config, exchangeName string) error {
	err := b.QueueBind(broker.QueueBind{
		Name:     reviewsQ,
		Key:      cfg.String("rabbitmq.reviews_routing_key", "1"),
		Exchange: exchangeName,
	}, broker.QueueBind{
		Name:     gamesQ,
		Key:      cfg.String("rabbitmq.games_routing_key", "2"),
		Exchange: exchangeName,
	})

	if err != nil {
		b.Close()
		return err
	}

	return nil
}
