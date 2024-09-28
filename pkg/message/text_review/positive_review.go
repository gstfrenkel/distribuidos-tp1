package text_review

import (
	"bytes"
	"encoding/gob"

	msg "tp1/pkg/message"
	"tp1/pkg/message/utils"
)

type message map[utils.Key][]string

func New(m map[utils.Key][]string) msg.Message {
	return message(m)
}

func FromBytes(b []byte) (msg.Message, error) {
	var m message
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return m, dec.Decode(&m)
}

func (m message) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m message) ToMessage(id msg.ID) (msg.Message, error) {
	return nil, nil
}
