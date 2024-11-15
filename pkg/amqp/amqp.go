package amqp

import (
	"strconv"
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
	Query1originId
	Query2originId
	Query3originId
	Query4originId
	Query5originId
)

var EmptyEof, _ = message.Eof{}.ToBytes()

type Header struct {
	SequenceId string
	ClientId   string
	OriginId   uint8
	MessageId  message.ID
}

type Delivery = amqp.Delivery
type Publishing = amqp.Publishing
type Queue amqp.Queue

type QueueBind struct {
	Exchange string
	Name     string
	Key      string
}

type DestinationEof Destination

type Exchange struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type Destination struct {
	Exchange  string `json:"exchange"`  // Exchange name.
	Key       string `json:"key"`       // Routing key format. MUST contain "%d" at the end if Consumers > 0.
	Name      string `json:"name"`      // Queue name format. MUST contain "%d" at the end if Consumers > 0.
	Consumers uint8  `json:"consumers"` // May be 0 if the number of queues does not scale up with the number of consumer workers.
}

type MessageBroker interface {
	QueueDeclare(name ...string) ([]Queue, error)
	ExchangeDeclare(exchange ...Exchange) error
	QueueBind(binds ...QueueBind) error
	ExchangeBind(dst, key, src string) error
	Publish(exchange, key string, msg []byte, headers map[string]any) error
	Consume(queue, consumer string, autoAck, exclusive bool) (<-chan Delivery, error)
	HandleEofMessage(workerId, peers uint8, msg []byte, headers map[string]any, input DestinationEof, outputs ...DestinationEof) error
	Close()
}

func HeadersFromDelivery(delivery Delivery) Header {
	originId, ok := delivery.Headers[OriginIdHeader]
	if !ok {
		originId = 255
	}

	sequenceId, ok := delivery.Headers[SequenceIdHeader]
	if !ok || sequenceId == nil {
		sequenceId = "0-0"
	}

	return Header{
		MessageId:  message.ID(delivery.Headers[MessageIdHeader].(uint8)),
		OriginId:   uint8(originId.(int)),
		ClientId:   delivery.Headers[ClientIdHeader].(string),
		SequenceId: sequenceId.(string),
	}
}

func (h Header) ToString() []string {
	return []string{
		h.SequenceId,
		h.ClientId,
		strconv.Itoa(int(h.OriginId)),
		strconv.Itoa(int(h.MessageId)),
	}
}
