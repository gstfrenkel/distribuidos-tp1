package message

type GameNames []GameName

type GameName struct {
	GameId   int64
	GameName string
}

func FromBytes(b []byte) (GameName, error) {
	var m GameName
	return m, fromBytes(b, &m)
}

func (g GameNames) ToBytes() ([]byte, error) {
	return toBytes(g)
}

func (g GameName) ToBytes() ([]byte, error) {
	return toBytes(g)
}
