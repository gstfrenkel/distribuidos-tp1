package messages

type Message interface {
	ToBytes() ([]byte, error)
	FromBytes([]byte) error
}

type MessageId uint8

const (
	REVIEW_ID_MSG MessageId = iota + 1
	GAME_ID_MSG
	EOF_MSG
)
