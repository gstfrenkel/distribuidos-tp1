package broker

import (
	"tp1/pkg/amqp"
	"tp1/pkg/message"

	amqpgo "github.com/rabbitmq/amqp091-go"
)

type messageBroker struct {
	conn *amqpgo.Connection
	ch   *amqpgo.Channel
}

// NewBroker creates a new AMQP connection and channel
func NewBroker() (amqp.MessageBroker, error) {
	conn, err := amqpgo.Dial("amqp://guest:guest@rabbitmq:5672/")
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

func (b *messageBroker) HandleEofMessage(workerId, peers uint8, msg []byte, headers map[string]any, input amqp.DestinationEof, outputs ...amqp.DestinationEof) error {
	workersVisited, err := message.EofFromBytes(msg)
	if err != nil {
		return err
	}

	if !workersVisited.Contains(workerId) {
		workersVisited = append(workersVisited, workerId)
	}

	if headers == nil {
		headers = map[string]any{amqp.MessageIdHeader: message.EofMsg}
	} else {
		headers[amqp.MessageIdHeader] = message.EofMsg
	}

	if uint8(len(workersVisited)) < peers {
		bytes, err := workersVisited.ToBytes()
		if err != nil {
			return err
		}
		return b.Publish(input.Exchange, input.Key, bytes, headers)
	}

	for _, output := range outputs {
		if err = b.Publish(output.Exchange, output.Key, message.Eof{}, headers); err != nil {
			return err
		}
	}
	return nil
}

func (b *messageBroker) Close() {
	_ = b.conn.Close()
	_ = b.ch.Close()
}
