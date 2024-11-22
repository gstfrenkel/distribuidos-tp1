package rabbit

import (
	"tp1/pkg/amqp"
	"tp1/pkg/config"
)

const (
	gamesPlatformQKey     = "rabbitmq.games_platform_q"
	gamesActionQKey       = "rabbitmq.games_action_q"
	gamesIndieQKey        = "rabbitmq.games_indie_q"
	reportsKey            = "rabbitmq.reports"
	reviewsQKey           = "rabbitmq.reviews_q"
	reviewsQDefault       = "reviews"
	gamesPlatQDefault     = "games_platform"
	gamesActionQDefault   = "games_action"
	gamesIndieQDefault    = "games_indie"
	reportsDefault        = "reports"
	exchangeKindKey       = "rabbitmq.exchange_kind"
	exKindDefault         = "direct"
	reviewsRoutingKey     = "rabbitmq.reviews_routing_key"
	reviewsRoutingDefault = "review"
	gamesRoutingKey       = "rabbitmq.games_routing_key"
	gamesRoutingDefault   = "game"
)

func CreateGatewayQueues(b amqp.MessageBroker, cfg config.Config) ([]amqp.Queue, error) {
	reviewsAndGamesQ, err := b.QueueDeclare(cfg.String(reviewsQKey, reviewsQDefault),
		cfg.String(gamesPlatformQKey, gamesPlatQDefault),
		cfg.String(gamesActionQKey, gamesActionQDefault),
		cfg.String(gamesIndieQKey, gamesIndieQDefault),
		cfg.String(reportsKey, reportsDefault))
	if err != nil {
		b.Close()
		return []amqp.Queue{}, err
	}

	return reviewsAndGamesQ, nil
}

func CreateExchange(cfg config.Config, b amqp.MessageBroker, exchangeName string) (string, error) {
	err := b.ExchangeDeclare(amqp.Exchange{Name: exchangeName, Kind: cfg.String(exchangeKindKey, exKindDefault)})
	if err != nil {
		b.Close()
		return "", err
	}
	return exchangeName, nil
}

func BindGatewayQueuesToExchange(b amqp.MessageBroker, queues []amqp.Queue, cfg config.Config, exchangeName string, reportsExchangeName string) error {
	gamesKey := cfg.String(gamesRoutingKey, gamesRoutingDefault)
	err := b.QueueBind(
		amqp.QueueBind{
			Name:     queues[0].Name,
			Key:      cfg.String(reviewsRoutingKey, reviewsRoutingDefault),
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
		}, amqp.QueueBind{
			Name:     queues[4].Name,
			Exchange: reportsExchangeName,
		},
	)

	if err != nil {
		b.Close()
		return err
	}

	return nil
}
