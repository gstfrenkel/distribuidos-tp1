package amqpconn

import (
	"tp1/pkg/broker"
	"tp1/pkg/message"

	amqp "github.com/rabbitmq/amqp091-go"
)

const MessageIdHeader = "x-message-id"

type Delivery = amqp.Delivery
type Publishing = amqp.Publishing

type messageBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewBroker creates a new AMQP connection and channel
func NewBroker() (broker.MessageBroker, error) {
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

// QueueDeclare declares new queues
func (b *messageBroker) QueueDeclare(names ...string) ([]broker.Queue, error) {
	var queues []broker.Queue

	for _, n := range names {
		q, err := b.ch.QueueDeclare(n, true, false, false, false, nil)
		if err != nil {
			return nil, err
		} else {
			queues = append(queues, broker.Queue(q))
		}
	}

	return queues, nil
}

// ExchangeDeclare declares new exchanges.
func (b *messageBroker) ExchangeDeclare(exchange ...broker.Exchange) error {
	for _, ex := range exchange {
		if err := b.ch.ExchangeDeclare(ex.Name, ex.Kind, true, false, false, false, nil); err != nil {
			return err
		}
	}
	return nil
}

// QueueBind binds queues to their respective exchanges.
func (b *messageBroker) QueueBind(binds ...broker.QueueBind) error {
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

func (b *messageBroker) HandleEofMessage(workerId, peers uint8, msg []byte, input broker.DestinationEof, outputs ...broker.DestinationEof) error {
	workersVisited, err := message.EofFromBytes(msg)
	if err != nil {
		return err
	}

	if !workersVisited.Contains(workerId) {
		workersVisited = append(workersVisited, workerId)
	}

	if uint8(len(workersVisited)) < peers {
		bytes, err := workersVisited.ToBytes()
		if err != nil {
			return err
		}
		return b.Publish(input.Exchange, input.Key, uint8(message.EofMsg), bytes)
	}

	for _, output := range outputs {
		if err = b.Publish(output.Exchange, output.Key, uint8(message.EofMsg), message.Eof{}); err != nil {
			return err
		}
	}
	return nil
}

func (b *messageBroker) Close() {
	_ = b.conn.Close()
	_ = b.ch.Close()
}

func publishingFromBytes(msgId uint8, msg []byte) Publishing {
	return amqp.Publishing{
		ContentType: "application/octet-stream",
		Headers:     map[string]interface{}{MessageIdHeader: msgId},
		Body:        msg,
	}
}
