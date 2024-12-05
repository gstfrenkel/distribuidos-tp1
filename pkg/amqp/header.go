package amqp

import (
	"errors"
	"strconv"
	"tp1/pkg/sequence"

	"tp1/pkg/message"
)

const (
	defaultOriginId = 255

	sequenceIdIdx = 0
	clientIdIdx   = 1
	originIdIdx   = 2
	messageIdIdx  = 3

	HeaderLen = 4
)

// Header represents a RabbitMQ message's header.
type Header struct {
	SequenceId string
	ClientId   string
	OriginId   uint8
	MessageId  message.Id
}

// HeadersFromDelivery creates a new Header from a Delivery.
func HeadersFromDelivery(delivery Delivery) Header {
	originId, ok := delivery.Headers[OriginIdHeader]
	if !ok || originId == nil {
		originId = defaultOriginId
	}

	sequenceId, ok := delivery.Headers[SequenceIdHeader]
	if !ok || sequenceId == nil {
		sequenceId = "0-0"
	}

	return Header{
		MessageId:  message.Id(delivery.Headers[MessageIdHeader].(uint8)),
		OriginId:   originId.(uint8),
		ClientId:   delivery.Headers[ClientIdHeader].(string),
		SequenceId: sequenceId.(string),
	}
}

// HeaderFromStrings creates a new Header from a human-readable slice of strings.
func HeaderFromStrings(header []string) (*Header, error) {
	if len(header) < HeaderLen {
		return nil, errors.New("not enough arguments")
	}

	originId, err := strconv.Atoi(header[originIdIdx])
	if err != nil {
		return nil, err
	}

	messageId, err := strconv.Atoi(header[messageIdIdx])
	if err != nil {
		return nil, err
	}

	return &Header{
		SequenceId: header[sequenceIdIdx],
		ClientId:   header[clientIdIdx],
		OriginId:   uint8(originId),
		MessageId:  message.Id(messageId),
	}, nil
}

// WithSequenceId sets a new sequence Id and returns the updated Header.
func (h Header) WithSequenceId(sequenceId sequence.Source) Header {
	h.SequenceId = sequenceId.ToString()
	return h
}

// WithOriginId sets a new origin Id and returns the updated Header.
func (h Header) WithOriginId(originId uint8) Header {
	h.OriginId = originId
	return h
}

// WithMessageId sets a new message Id and returns the updated Header.
func (h Header) WithMessageId(messageId message.Id) Header {
	h.MessageId = messageId
	return h
}

// ToMap turns the Header into a delivery ready map.
func (h Header) ToMap() map[string]any {
	return map[string]any{
		SequenceIdHeader: h.SequenceId,
		ClientIdHeader:   h.ClientId,
		OriginIdHeader:   h.OriginId,
		MessageIdHeader:  uint8(h.MessageId),
	}
}

// ToString turns the Header into human-readable slice of strings.
func (h Header) ToString() []string {
	return []string{
		h.SequenceId,
		h.ClientId,
		strconv.Itoa(int(h.OriginId)),
		strconv.Itoa(int(h.MessageId)),
	}
}
