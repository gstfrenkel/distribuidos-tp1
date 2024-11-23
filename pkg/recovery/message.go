package recovery

import "tp1/pkg/amqp"

type Message interface {
	Header() amqp.Header
	Message() []byte
}

type message struct {
	record Record
}

func NewMessage(record Record) Message {
	return &message{record: record}
}

func (m *message) Header() amqp.Header {
	return m.record.Header()
}

func (m *message) Message() []byte {
	return m.record.Message()
}
