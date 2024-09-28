package message

import "errors"

var ErrFailedToConvert = errors.New("failed to convert to message with ID %d")

const (
	ReviewIdMsg ID = iota + 1
	GameIdMsg
	EofMsg

	PositiveReviewID
	NegativeReviewID
	PositiveReviewWithTextID
)

type ID uint8

type Message interface {
	ToBytes() ([]byte, error)
	ToMessage(id ID) (Message, error)
}
