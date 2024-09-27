package amqpconn

import (
	"tp1/pkg/broker"

	amqp "github.com/rabbitmq/amqp091-go"
)

type messageBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

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

func (b *messageBroker) QueueDeclare(name string) (amqp.Queue, error) {
	return b.ch.QueueDeclare(name, true, false, false, false, nil)
}

func (b *messageBroker) ExchangeDeclare(name string, kind string) error {
	return b.ch.ExchangeDeclare(name, kind, true, false, false, false, nil)
}

func (b *messageBroker) QueueBind(name string, key string, exchange string) error {
	return b.ch.QueueBind(name, key, exchange, false, nil)
}

func (b *messageBroker) ExchangeBind(dst string, key string, src string) error {
	return b.ch.ExchangeBind(dst, key, src, false, nil)
}

func (b *messageBroker) Publish(exchange string, key string, msg amqp.Publishing) error {
	return b.ch.Publish(exchange, key, true, false, msg)
}

func (b *messageBroker) Close() {
	_ = b.conn.Close()
	_ = b.ch.Close()
}
