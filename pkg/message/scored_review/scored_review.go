package scored_review

import (
	"bytes"
	"encoding/gob"
)

type Messages []Message

type Message struct {
	GameId   int64
	GameName string
	Votes    int64
}

func FromBytes(b []byte) (Message, error) {
	var m Message
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return m, dec.Decode(&m)
}

func (m Message) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
