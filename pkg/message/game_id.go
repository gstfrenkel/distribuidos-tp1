package message

type GameId int64

func NewGameId(id int64) GameId {
	return NewGameId(id)
}

func FromBytes(b []byte) (GameId, error) {
	var m GameId
	return m, fromBytes(b, &m)
}

func (g GameId) ToBytes() ([]byte, error) {
	return toBytes(g)
}
