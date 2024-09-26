package amqp

import amqp "github.com/rabbitmq/amqp091-go"

type MessageBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func New(name string) (*MessageBroker, error) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		conn.Close()
		ch.Close()
		return nil, err
	}

	return &MessageBroker{
		conn: conn,
		ch:   ch,
		q:    q,
	}, nil
}

func (b *MessageBroker) Close() {
	b.conn.Close()
	b.ch.Close()
}
