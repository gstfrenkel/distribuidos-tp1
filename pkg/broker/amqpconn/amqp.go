package amqpconn

import (
	"tp1/pkg/broker"

	amqp "github.com/rabbitmq/amqp091-go"
)

const MessageIdHeader = "x-message-id"

type Delivery = amqp.Delivery
type Publishing = amqp.Publishing

type messageBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// New creates a new AMQP connection and channel
func New() (broker.MessageBroker, error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &messageBroker{
		conn: conn,
		ch:   ch,
	}, nil
}

// QueueDeclare declares a new queue
func (b *messageBroker) QueueDeclare(name string) (amqp.Queue, error) {
	return b.ch.QueueDeclare(name, true, false, false, false, nil)
}

// QueuesDeclare declares new queues
func (b *messageBroker) QueuesDeclare(names ...string) ([]amqp.Queue, error) {
	var queues []amqp.Queue

	for _, n := range names {
		q, err := b.ch.QueueDeclare(n, true, false, false, false, nil)
		if err != nil {
			return nil, err
		} else {
			queues = append(queues, q)
		}
	}

	return queues, nil
}

// ExchangeDeclare declares a new exchange
func (b *messageBroker) ExchangeDeclare(name, kind string) error {
	return b.ch.ExchangeDeclare(name, kind, true, false, false, false, nil)
}

// ExchangesDeclare declares new exchanges.
func (b *messageBroker) ExchangesDeclare(exchanges ...broker.Exchange) error {
	for _, ex := range exchanges {
		if err := b.ExchangeDeclare(ex.Name, ex.Kind); err != nil {
			return err
		}
	}
	return nil
}

// QueueBind binds a queue to an exchange
func (b *messageBroker) QueueBind(name, key, exchange string) error {
	return b.ch.QueueBind(name, key, exchange, false, nil)
}

// QueuesBind binds queues to their respective exchanges.
func (b *messageBroker) QueuesBind(binds ...broker.QueueBind) error {
	for _, bind := range binds {
		if err := b.ch.QueueBind(bind.Name, bind.Key, bind.Exchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}

// ExchangeBind binds an exchange to another exchange
func (b *messageBroker) ExchangeBind(dst, key, src string) error {
	return b.ch.ExchangeBind(dst, key, src, false, nil)
}

// Publish sends a message to an exchange
func (b *messageBroker) Publish(exchange, key string, msgId uint8, msg []byte) error {
	return b.ch.Publish(exchange, key, true, false, publishingFromBytes(msgId, msg))
}

func (b *messageBroker) Consume(queue, consumer string, autoAck, exclusive bool) (<-chan Delivery, error) {
	return b.ch.Consume(queue, consumer, autoAck, exclusive, false, false, nil)
}

func publishingFromBytes(msgId uint8, msg []byte) Publishing {
	return amqp.Publishing{
		ContentType: "application/octet-stream",
		Headers:     map[string]interface{}{MessageIdHeader: msgId},
		Body:        msg,
	}
}

func (b *messageBroker) Close() {
	_ = b.conn.Close()
	_ = b.ch.Close()
}
