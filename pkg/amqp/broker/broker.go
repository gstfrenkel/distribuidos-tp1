package broker

import (
	"tp1/pkg/amqp"

	amqpgo "github.com/rabbitmq/amqp091-go"
)

type messageBroker struct {
	conn *amqpgo.Connection
	ch   *amqpgo.Channel
}

// NewBroker creates a new AMQP connection and channel
func NewBroker() (amqp.MessageBroker, error) {
	retries := 0
	var conn *amqpgo.Connection
	var err error

	for retries < 3 {
		conn, err = amqpgo.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			break
		}
		retries++
	}

	if retries == 3 {
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

// QueueDeclare declares new queues
func (b *messageBroker) QueueDeclare(names ...string) ([]amqp.Queue, error) {
	var queues []amqp.Queue

	for _, n := range names {
		q, err := b.ch.QueueDeclare(n, true, false, false, false, nil)
		if err != nil {
			return nil, err
		} else {
			queues = append(queues, amqp.Queue(q))
		}
	}

	return queues, nil
}

// ExchangeDeclare declares new exchanges.
func (b *messageBroker) ExchangeDeclare(exchange ...amqp.Exchange) error {
	for _, ex := range exchange {
		if err := b.ch.ExchangeDeclare(ex.Name, ex.Kind, true, false, false, false, nil); err != nil {
			return err
		}
	}
	return nil
}

// QueueBind binds queues to their respective exchanges.
func (b *messageBroker) QueueBind(binds ...amqp.QueueBind) error {
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
func (b *messageBroker) Publish(exchange, key string, msg []byte, headers map[string]any) error {
	return b.ch.Publish(exchange, key, true, false, amqp.Publishing{
		ContentType: "application/octet-stream",
		Headers:     headers,
		Body:        msg,
	})
}

func (b *messageBroker) Consume(queue, consumer string, autoAck, exclusive bool) (<-chan amqp.Delivery, error) {
	return b.ch.Consume(queue, consumer, autoAck, exclusive, false, false, nil)
}

func (b *messageBroker) Close() {
	_ = b.conn.Close()
	_ = b.ch.Close()
}
