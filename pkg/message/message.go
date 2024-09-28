package message

import (
	"errors"
)

var ErrFailedToConvert = errors.New("failed to convert to message with ID %d")
var ErrEmptyByteSlice = errors.New("unable to convert empty slice of bytes into message")

const (
	ReviewIdMsg ID = iota + 1
	GameIdMsg
	EofMsg

	ReviewID
	PositiveReviewID
	NegativeReviewID
	PositiveReviewWithTextID
)

type ID uint8

/*type Message interface {
	ToBytes() ([]byte, error)
	ToMessage(id ID) ([]Message, error)
	GameId()
}*/
