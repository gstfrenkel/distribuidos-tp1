package message

const (
	ReviewIdMsg ID = iota + 1
	GameIdMsg
	EofMsg
)

type ID uint8

type Message interface {
	ToBytes() ([]byte, error)
}
