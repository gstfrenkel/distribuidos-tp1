package broker

import amqp "github.com/rabbitmq/amqp091-go"

type Queue amqp.Queue

type QueueBind struct {
	Exchange string
	Name     string
	Key      string
}

type Destination struct {
	Exchange string
	Key      string
}

type Aaaa struct {
	Exchange  string // Exchange name.
	Key       string // Routing key format. MUST contain "%d" at the end if Consumers > 0.
	Name      string // Queue name format. MUST contain "%d" at the end if Consumers > 0.
	Consumers uint8  // May be 0 if the number of queues does not scale up with the number of consumer workers.
}

type MessageBroker interface {
	QueueDeclare(name ...string) ([]Queue, error)
	ExchangeDeclare(exchanges map[string]string) error
	QueueBind(binds ...QueueBind) error
	ExchangeBind(dst, key, src string) error
	Publish(exchange, key string, msgId uint8, msg []byte) error
	Consume(queue, consumer string, autoAck, exclusive bool) (<-chan amqp.Delivery, error)
	HandleEofMessage(workerId, peers uint8, msg []byte, input Destination, outputs ...Destination) error
	Close()
}
