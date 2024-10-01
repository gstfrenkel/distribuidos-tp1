package rabbit

import (
	"github.com/rabbitmq/amqp091-go"
	"tp1/pkg/broker"
	"tp1/pkg/config"
)

func CreateGatewayQueues(b broker.MessageBroker, cfg config.Config) ([]amqp091.Queue, error) {
	reviewsAndGamesQ, err := b.QueueDeclare(cfg.String("rabbitmq.reviews_q", "reviews"),
		cfg.String("rabbitmq.games_platform_q", "games_platform"),
		cfg.String("rabbitmq.games_shooter_q", "games_shooter"),
		cfg.String("rabbitmq.games_indie_q", "games_indie"))
	if err != nil {
		b.Close()
		return []amqp091.Queue{}, err
	}

	return reviewsAndGamesQ, nil
}

func CreateGatewayExchange(cfg config.Config, b broker.MessageBroker) (string, error) {
	exchangeName := cfg.String("rabbitmq.exchange_name", "gateway")
	err := b.ExchangeDeclare(broker.Exchange{Name: exchangeName, Kind: cfg.String("rabbitmq.exchange_type", "direct")})
	if err != nil {
		b.Close()
		return "", err
	}
	return exchangeName, nil
}

func BindGatewayQueuesToExchange(b broker.MessageBroker, queues []amqp091.Queue, cfg config.Config, exchangeName string) error {
	gamesKey := cfg.String("rabbitmq.games_routing_key", "game")
	err := b.QueueBind(broker.QueueBind{
		Name:     queues[0].Name,
		Key:      cfg.String("rabbitmq.reviews_routing_key", "review"),
		Exchange: exchangeName,
	}, broker.QueueBind{
		Name:     queues[1].Name,
		Key:      gamesKey,
		Exchange: exchangeName,
	}, broker.QueueBind{
		Name:     queues[2].Name,
		Key:      gamesKey,
		Exchange: exchangeName,
	}, broker.QueueBind{
		Name:     queues[3].Name,
		Key:      gamesKey,
		Exchange: exchangeName,
	})

	if err != nil {
		b.Close()
		return err
	}

	return nil
}
