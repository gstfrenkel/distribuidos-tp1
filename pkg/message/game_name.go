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

func GameNameFromAnyToBytes(m []any) ([]byte, error) {
	var g GameNames
	for _, gameName := range m {
		g = append(g, gameName.(GameName))
	}
	return toBytes(g)
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

func ToQ4ResultString(results string) string {
	header := fmt.Sprintf("Q4: Juegos del género Action con más de 5.000 reseñas negativas en idioma inglés\n")
	return header + results
}

func (names GameNames) ToStringAux() string {
	var gamesInfo []string
	for _, game := range names {
		gamesInfo = append(gamesInfo, fmt.Sprintf("Juego: [%s], ID: [%d] \n", game.GameName, game.GameId))
	}
	return strings.Join(gamesInfo, "")
}
