package rabbit

import (
	"fmt"

	"tp1/pkg/amqp"
	"tp1/pkg/config"
)

const (
	minSize = 5

	queue     = "queue"
	key       = "key"
	consumers = "consumers"

	reviewsKey  = "rabbitmq.reviews"
	actionKey   = "rabbitmq.action"
	indieKey    = "rabbitmq.indie"
	platformKey = "rabbitmq.platform"
	reportsKey  = "rabbitmq.reports"

	exchangeNameKey = "rabbitmq.exchange_name"
	exchangeKindKey = "rabbitmq.exchange_kind"
	exKindDefault   = "direct"
)

var queueKeys = []string{
	reviewsKey,
	actionKey,
	indieKey,
	platformKey,
}

func buildKey(prefix, suffix string) string {
	return fmt.Sprintf("%s.%s", prefix, suffix)
}

func buildQueue(queuePrefix, keyPrefix string, id uint8) (string, string) {
	return fmt.Sprintf(queuePrefix, id), fmt.Sprintf(keyPrefix, id)
}

func buildQueues(queuePrefix, keyPrefix string, consumers uint8) ([]string, []string) {
	names := make([]string, 0, consumers)
	keys := make([]string, 0, consumers)

	for i := uint8(0); i < consumers; i++ {
		name, k := buildQueue(queuePrefix, keyPrefix, i)
		names = append(names, name)
		keys = append(keys, k)
	}

	return names, keys
}

func CreateGatewayQueues(id uint8, b amqp.MessageBroker, cfg config.Config) ([]amqp.Destination, []amqp.Queue, error) {
	exchange := cfg.String(exchangeNameKey, "")
	if err := createExchange(b, exchange, cfg.String(exchangeKindKey, exKindDefault)); err != nil {
		return nil, nil, err
	}

	destinations := make([]amqp.Destination, 0, minSize)
	names := make([]string, 0, minSize)
	binds := make([]amqp.QueueBind, 0, minSize)

	for _, k := range queueKeys {
		queueKey := cfg.String(buildKey(k, key), "")
		queueConsumers := cfg.Uint8(buildKey(k, consumers), 1)

		destinations = append(destinations, amqp.Destination{Exchange: exchange, Key: queueKey, Consumers: queueConsumers})

		name, keys := buildQueues(cfg.String(buildKey(k, queue), ""), queueKey, queueConsumers)
		names = append(names, name...)
		for i := 0; i < len(keys); i++ {
			binds = append(binds, amqp.QueueBind{Exchange: exchange, Name: name[i], Key: keys[i]})
		}
	}

	name, _ := buildQueue(cfg.String(buildKey(reportsKey, queue), ""), "", id)
	names = append(names, name)

	queues, err := b.QueueDeclare(names...)
	if err != nil {
		b.Close()
		return nil, nil, err
	}

	if err = b.QueueBind(binds...); err != nil {
		return nil, nil, err
	}

	return destinations, queues, nil
}

func createExchange(b amqp.MessageBroker, name, kind string) error {
	err := b.ExchangeDeclare(amqp.Exchange{Name: name, Kind: kind})
	if err != nil {
		b.Close()
		return err
	}
	return nil
}
