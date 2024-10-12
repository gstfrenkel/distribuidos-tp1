package rabbit

import (
	"tp1/pkg/amqp"
	"tp1/pkg/config"
)

func CreateGatewayQueues(b amqp.MessageBroker, cfg config.Config) ([]amqp.Queue, error) {
	reviewsAndGamesQ, err := b.QueueDeclare(cfg.String("rabbitmq.reviews_q", "reviews"),
		cfg.String("rabbitmq.games_platform_q", "games_platform"),
		cfg.String("rabbitmq.games_action_q", "games_action"),
		cfg.String("rabbitmq.games_indie_q", "games_indie"))
	if err != nil {
		b.Close()
		return []amqp.Queue{}, err
	}

	return reviewsAndGamesQ, nil
}

// func CreateGatewayExchange(cfg config.Config, b amqp.MessageBroker) (string, error) {
// 	exchangeName := cfg.String("rabbitmq.exchange_name", "gateway")
// 	err := b.ExchangeDeclare(amqp.Exchange{Name: exchangeName, Kind: cfg.String("rabbitmq.exchange_kind", "direct")})
// 	if err != nil {
// 		b.Close()
// 		return "", err
// 	}
// 	return exchangeName, nil
// }

// func CreateReportsExchange(cfg config.Config, b amqp.MessageBroker) (string, error) {
// 	exchangeName := cfg.String("rabbitmq.exchange_name", "gateway")
// 	err := b.ExchangeDeclare(amqp.Exchange{Name: exchangeName, Kind: cfg.String("rabbitmq.exchange_kind", "direct")})
// 	if err != nil {
// 		b.Close()
// 		return "", err
// 	}
// 	return exchangeName, nil
// }

func CreateExchange(cfg config.Config, b amqp.MessageBroker, exchangeName string) (string, error) {
	err := b.ExchangeDeclare(amqp.Exchange{Name: exchangeName, Kind: cfg.String("rabbitmq.exchange_kind", "direct")})
	if err != nil {
		b.Close()
		return "", err
	}
	return exchangeName, nil
}

func BindGatewayQueuesToExchange(b amqp.MessageBroker, queues []amqp.Queue, cfg config.Config, exchangeName string) error {
	gamesKey := cfg.String("rabbitmq.games_routing_key", "game")
	err := b.QueueBind(amqp.QueueBind{
		Name:     queues[0].Name,
		Key:      cfg.String("rabbitmq.reviews_routing_key", "review"),
		Exchange: exchangeName,
	}, amqp.QueueBind{
		Name:     queues[1].Name,
		Key:      gamesKey,
		Exchange: exchangeName,
	}, amqp.QueueBind{
		Name:     queues[2].Name,
		Key:      gamesKey,
		Exchange: exchangeName,
	}, amqp.QueueBind{
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

func BindReportsQueueToExchange(b amqp.MessageBroker, cfg config.Config, exchangeName string) error {
	reportsQueue := cfg.String("rabbitmq.reports", "reports")
	err := b.QueueBind(amqp.QueueBind{
		Name:     reportsQueue,
		Exchange: exchangeName,
	})

	if err != nil {
		b.Close()
		return err
	}

	return nil
}
