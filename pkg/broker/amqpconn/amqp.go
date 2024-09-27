package amqpconn

import (
	"tp1/pkg/broker"

	amqp "github.com/rabbitmq/amqp091-go"
)

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

// ExchangeDeclare declares a new exchange
func (b *messageBroker) ExchangeDeclare(name string, kind string) error {
	return b.ch.ExchangeDeclare(name, kind, true, false, false, false, nil)
}

// QueueBind binds a queue to an exchange
func (b *messageBroker) QueueBind(name string, key string, exchange string) error {
	return b.ch.QueueBind(name, key, exchange, false, nil)
}

// ExchangeBind binds an exchange to another exchange
func (b *messageBroker) ExchangeBind(dst string, key string, src string) error {
	return b.ch.ExchangeBind(dst, key, src, false, nil)
}

// Publish sends a message to an exchange
func (b *messageBroker) Publish(exchange string, key string, msg []byte) error {
	return b.ch.Publish(exchange, key, true, false, publishingFromBytes(msg))
}

func publishingFromBytes(msg []byte) amqp.Publishing {
	return amqp.Publishing{
		ContentType: "application/octet-stream",
		Body:        msg,
	}
}

func (b *messageBroker) Close() {
	_ = b.conn.Close()
	_ = b.ch.Close()
}
