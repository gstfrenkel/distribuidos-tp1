package broker

import amqp "github.com/rabbitmq/amqp091-go"

type Exchange struct {
	Name string
	Kind string
}

type QueueBind struct {
	Name     string
	Key      string
	Exchange string
}

type MessageBroker interface {
	QueueDeclare(name string) (amqp.Queue, error)
	QueuesDeclare(name ...string) ([]amqp.Queue, error)
	ExchangeDeclare(name string, kind string) error
	ExchangesDeclare(exchange ...Exchange) error
	QueueBind(name, key, exchange string) error
	QueuesBind(binds ...QueueBind) error
	ExchangeBind(dst, key, src string) error
	Publish(exchange, key string, msgId uint8, msg []byte) error
	Consume(queue, consumer string, autoAck, exclusive bool) (<-chan amqp.Delivery, error)
	Close()
}
