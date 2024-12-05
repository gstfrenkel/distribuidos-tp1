package message

import (
	"fmt"
	"strings"
)

type GameNames []GameName

type GameName struct {
	GameId   int64
	GameName string
}

func GameNameFromBytes(b []byte) (GameName, error) {
	var m GameName
	return m, fromBytes(b, &m)
}

func GameNamesFromBytes(b []byte) (GameNames, error) {
	var m GameNames
	return m, fromBytes(b, &m)
}

func (g GameNames) ToBytes() ([]byte, error) {
	return toBytes(g)
}

func (g GameName) ToBytes() ([]byte, error) {
	return toBytes(g)
}

func ToQ4ResultString(results string) string {
	header := fmt.Sprintf("Q4:\n")
	return header + results
}

func (g GameNames) ToStringAux() string {
	var gamesInfo []string
	for _, info := range g {
		gamesInfo = append(gamesInfo, fmt.Sprintf("Juego: [%s], Id: [%d]", info.GameName, info.GameId))
	}
	return strings.Join(gamesInfo, "\n")
}
