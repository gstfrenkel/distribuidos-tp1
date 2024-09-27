package messages

type Message interface {
	ToBytes() ([]byte, error)
	FromBytes([]byte) error
}
