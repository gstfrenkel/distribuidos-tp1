package broker

import amqp "github.com/rabbitmq/amqp091-go"

type Exchange struct {
	Name string
	Kind string
}

type Destination struct {
	Exchange string
	Key      string
}

type QueueBind struct {
	Name     string
	Key      string
	Exchange string
}

type MessageBroker interface {
	QueueDeclare(name ...string) ([]amqp.Queue, error)
	ExchangeDeclare(exchange ...Exchange) error
	QueueBind(binds ...QueueBind) error
	ExchangeBind(dst, key, src string) error
	Publish(exchange, key string, msgId uint8, msg []byte) error
	Consume(queue, consumer string, autoAck, exclusive bool) (<-chan amqp.Delivery, error)
	HandleEofMessage(workerId uint8, peers uint8, msg []byte, input Destination, outputs ...Destination) error
	Close()
}
