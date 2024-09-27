package broker

import amqp "github.com/rabbitmq/amqp091-go"

type MessageBroker interface {
	QueueDeclare(name string) (amqp.Queue, error)
	ExchangeDeclare(name string, kind string) error
	QueueBind(name string, key string, exchange string) error
	ExchangeBind(dst string, key string, src string) error
	Publish(exchange string, key string, msg []byte) error
	Close()
}
