package amqp

import (
	"errors"
	"strconv"

	"tp1/pkg/message"
)

type Header struct {
	SequenceId string
	ClientId   string
	OriginId   uint8
	MessageId  message.ID
}

func HeadersFromDelivery(delivery Delivery) Header {
	originId, ok := delivery.Headers[OriginIdHeader]
	if !ok {
		originId = 255
	}

	sequenceId, ok := delivery.Headers[SequenceIdHeader]
	if !ok || sequenceId == nil {
		sequenceId = "0-0"
	}

	return Header{
		MessageId:  message.ID(delivery.Headers[MessageIdHeader].(uint8)),
		OriginId:   uint8(originId.(int)),
		ClientId:   delivery.Headers[ClientIdHeader].(string),
		SequenceId: sequenceId.(string),
	}
}

func HeaderFromStrings(header []string) (*Header, error) {
	if len(header) < 4 {
		return nil, errors.New("not enough arguments")
	}

	originId, err := strconv.Atoi(header[2])
	if err != nil {
		return nil, err
	}

	messageId, err := strconv.Atoi(header[3])
	if err != nil {
		return nil, err
	}

	return &Header{
		SequenceId: header[0],
		ClientId:   header[1],
		OriginId:   uint8(originId),
		MessageId:  message.ID(messageId),
	}, nil
}

func (h Header) ToString() []string {
	return []string{
		h.SequenceId,
		h.ClientId,
		strconv.Itoa(int(h.OriginId)),
		strconv.Itoa(int(h.MessageId)),
	}
}
