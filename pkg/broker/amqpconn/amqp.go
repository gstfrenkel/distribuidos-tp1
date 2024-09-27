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

func (b *messageBroker) ExchangeDeclare(name string, kind string) error {
	return b.ch.ExchangeDeclare(name, kind, true, false, false, false, nil)
}

func (b *messageBroker) ExchangesDeclare(exchanges ...broker.Exchange) error {
	for _, ex := range exchanges {
		if err := b.ExchangeDeclare(ex.Name, ex.Kind); err != nil {
			return err
		}
	}
	return nil
}

func (b *messageBroker) QueueBind(name string, key string, exchange string) error {
	return b.ch.QueueBind(name, key, exchange, false, nil)
}

func (b *messageBroker) QueuesBind(binds ...broker.QueueBind) error {
	for _, bind := range binds {
		if err := b.ch.QueueBind(bind.Name, bind.Key, bind.Exchange, false, nil); err != nil {
			return err
		}
	}
	return nil
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
