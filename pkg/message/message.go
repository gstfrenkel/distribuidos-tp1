package message

import (
	"bytes"
	"encoding/gob"
	"errors"
)

var ErrFailedToConvert = errors.New("failed to convert to message with ID %d")
var ErrEmptyByteSlice = errors.New("unable to convert empty slice of bytes into message")

const (
	ReviewIdMsg ID = iota + 1
	GameIdMsg
	EofMsg

	ReviewID
	PositiveReviewID // ScoredReview message ID. Score is implicitly equal to 1.
	NegativeReviewID // ScoredReview message ID. Score is implicitly equal to -1.
	PositiveReviewWithTextID
	GameIdID // GameId message ID.
)

type ID uint8

func fromBytes(b []byte, msg any) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(b))
	return decoder.Decode(msg)
}

func toBytes(msg any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
