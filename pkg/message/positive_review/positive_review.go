package positive_review

import (
	msg "tp1/pkg/message"
	"tp1/pkg/message/utils"
)

type message map[utils.Key][]string

func New(m map[utils.Key][]string) msg.Message {
	return message(m)
}

func FromBytes() msg.Message {
	return message{}
}

func (m message) ToBytes() ([]byte, error) {
	return []byte{}, nil
}

func (m message) ToMessage(id msg.ID) (msg.Message, error) {
	return nil, nil
}
