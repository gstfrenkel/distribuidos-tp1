package message

import (
	"bytes"
	"fmt"
	"strings"
	"tp1/pkg/utils/encoding"
)

type GameNames []GameName

type GameName struct {
	GameId   int64
	GameName string
}

func GameNameFromBytes(b []byte) (GameName, error) {
	return GameNameFromBuf(bytes.NewBuffer(b))
}

func GameNamesFromBytes(b []byte) (GameNames, error) {
	var gs GameNames
	buf := bytes.NewBuffer(b)
	size, err := encoding.DecodeUint32(buf)
	if err != nil {
		return nil, err
	}

	for i := uint32(0); i < size; i++ {
		g, err := GameNameFromBuf(buf)
		if err != nil {
			return nil, err
		}

		gs = append(gs, g)
	}

	return gs, nil
}

func GameNameFromBuf(buf *bytes.Buffer) (GameName, error) {
	gId, err := encoding.DecodeInt64(buf)
	if err != nil {
		return GameName{}, err
	}

	gName, err := encoding.DecodeString(buf)
	if err != nil {
		return GameName{}, err
	}

	return GameName{GameId: gId, GameName: gName}, nil
}

func (g GameNames) ToBytes() ([]byte, error) {
	size := uint32(len(g))
	sizeBuff := bytes.Buffer{}
	if err := encoding.EncodeNumber(&sizeBuff, size); err != nil {
		return nil, err
	}

	b := make([]byte, 0, size)
	b = append(b, sizeBuff.Bytes()...)

	for _, gm := range g {
		gameB, err := gm.ToBytes()
		if err != nil {
			return nil, err
		}
		b = append(b, gameB...)
	}

	return b, nil
}

func (g GameName) ToBytes() ([]byte, error) {
	buff := bytes.Buffer{}
	if err := encoding.EncodeNumber(&buff, g.GameId); err != nil {
		return nil, err
	}

	if err := encoding.EncodeString(&buff, g.GameName); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
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
	return strings.Join(gamesInfo, "\n") + "\n"
}
