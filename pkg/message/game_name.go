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

func (names GameNames) ToResultString() string {
	header := fmt.Sprintf("Q4: Juegos del género Action con más de 5.000 reseñas negativas en idioma inglés\n")
	var namesInfo []string
	for _, name := range names {
		namesInfo = append(namesInfo, fmt.Sprintf("Juego: [%s] \n", name.GameName))
	}
	return header + strings.Join(namesInfo, "")
}
