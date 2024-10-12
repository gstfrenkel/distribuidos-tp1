package message

type GameNames []GameName

type GameName struct {
	GameId   int64
	GameName string
}

func GameNameFromBytes(b []byte) (GameName, error) {
	var m GameName
	return m, fromBytes(b, &m)
}

func (g GameNames) ToBytes() ([]byte, error) {
	return toBytes(g)
}

func ToBytes(m any) ([]byte, error) {
	return toBytes(m.(GameNames))
}

func (g GameName) ToBytes() ([]byte, error) {
	return toBytes(g)
}

func (g GameNames) ToAny() []any {
	var dataAny []any
	for _, gameName := range g {
		dataAny = append(dataAny, gameName)
	}
	return dataAny
}
