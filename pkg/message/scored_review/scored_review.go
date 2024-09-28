package scored_review

import (
	msg "tp1/pkg/message"
	"tp1/pkg/message/utils"
)

type message map[utils.Key]int64

func New(m map[utils.Key]int64) msg.Message {
	return message(m)
}

func FromBytes() msg.Message {

	return nil
}

func (m message) ToBytes() ([]byte, error) {
	return []byte{}, nil
}

func (m message) ToMessage(id msg.ID) (msg.Message, error) {
	return nil, nil
}
