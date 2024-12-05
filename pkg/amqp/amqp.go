package amqp

import (
	"tp1/pkg/message"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	MessageIdHeader  = "x-message-id"
	OriginIdHeader   = "x-origin-id"
	ClientIdHeader   = "x-client-id"
	SequenceIdHeader = "x-sequence-id"
)

const (
	ReviewOriginId uint8 = iota
	GameOriginId
	Query1OriginId
	Query2OriginId
	Query3OriginId
	Query4OriginId
	Query5OriginId
)

var EmptyEof, _ = message.Eof{}.ToBytes() // EmptyEof represents a serialized EOF message with no content.

type Delivery = amqp.Delivery
type Publishing = amqp.Publishing
type Queue amqp.Queue

// QueueBind represents the configuration needed to bind a queue to an exchange in RabbitMQ.
// It specifies the exchange to bind to, the name of the queue, and the routing key used for binding.
type QueueBind struct {
	Exchange string
	Name     string
	Key      string
}

type DestinationEof Destination

// Exchange represents the configuration of a RabbitMQ exchange.
type Exchange struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// Destination represents the configuration for routing messages to a specific exchange and queue.
// It supports flexible setups, including scaling with multiple consumer workers.
type Destination struct {
	Exchange  string `json:"exchange"`  // Exchange name.
	Key       string `json:"key"`       // Routing key format. MUST contain "%d" at the end if Consumers > 0.
	Name      string `json:"name"`      // Queue name format. MUST contain "%d" at the end if Consumers > 0.
	Consumers uint8  `json:"consumers"` // May be 0 if the number of queues does not scale up with the number of consumer workers.
	Single    bool   `json:"single"`
}

type MessageBroker interface {
	QueueDeclare(name ...string) ([]Queue, error)
	ExchangeDeclare(exchange ...Exchange) error
	QueueBind(binds ...QueueBind) error
	ExchangeBind(dst, key, src string) error
	Publish(exchange, key string, msg []byte, headers Header) error
	Consume(queue, consumer string, autoAck, exclusive bool) (<-chan Delivery, error)
	Close()
}
