package message

import (
	"bytes"
	"strings"
	"tp1/pkg/utils/encoding"
)

const sep = ","

type Game []game

type game struct {
	GameId          int64
	AveragePlaytime int64
	Name            string
	Genres          string
	ReleaseDate     string
	Windows         bool
	Mac             bool
	Linux           bool
}

func GamesFromBytes(b []byte) (Game, error) {
	var gs Game
	buff := bytes.NewBuffer(b)
	size, err := encoding.DecodeUint32(buff)
	if err != nil {
		return nil, err
	}

	for i := uint32(0); i < size; i++ {
		g, err := gameFromBytes(buff)
		if err != nil {
			return nil, err
		}

		gs = append(gs, g)
	}

	return gs, nil
}

func gameFromBytes(buff *bytes.Buffer) (game, error) {
	gameId, err := encoding.DecodeInt64(buff)
	if err != nil {
		return game{}, err
	}

	averagePlaytime, err := encoding.DecodeInt64(buff)
	if err != nil {
		return game{}, err
	}

	name, err := encoding.DecodeString(buff)
	if err != nil {
		return game{}, err
	}

	genres, err := encoding.DecodeString(buff)
	if err != nil {
		return game{}, err
	}

	releaseDate, err := encoding.DecodeString(buff)
	if err != nil {
		return game{}, err
	}

	windows, err := encoding.DecodeBool(buff)
	if err != nil {
		return game{}, err
	}

	mac, err := encoding.DecodeBool(buff)
	if err != nil {
		return game{}, err
	}

	linux, err := encoding.DecodeBool(buff)
	if err != nil {
		return game{}, err
	}

	return game{
		GameId:          gameId,
		AveragePlaytime: averagePlaytime,
		Name:            name,
		Genres:          genres,
		ReleaseDate:     releaseDate,
		Windows:         windows,
		Mac:             mac,
		Linux:           linux,
	}, nil
}

func GamesFromClientGames(clientGame []DataCSVGames) ([]byte, error) {
	gs := make(Game, 0, len(clientGame))
	for _, g := range clientGame {
		gs = append(gs, game{
			GameId:          g.AppID,
			Name:            g.Name,
			Windows:         g.Windows,
			Mac:             g.Mac,
			Linux:           g.Linux,
			Genres:          g.Genres,
			AveragePlaytime: g.AveragePlaytimeForever,
			ReleaseDate:     g.ReleaseDate,
		})
	}

	return gs.ToBytes()
}

func (g Game) ToBytes() ([]byte, error) {
	size := uint32(len(g))
	sizeBuff := bytes.Buffer{}
	if err := encoding.EncodeNumber(&sizeBuff, size); err != nil {
		return nil, err
	}

	b := make([]byte, 0, size)
	b = append(b, sizeBuff.Bytes()...)

	for _, gm := range g {
		gameB, err := gm.gameToBytes()
		if err != nil {
			return nil, err
		}
		b = append(b, gameB...)
	}

	return b, nil
}

func (g game) gameToBytes() ([]byte, error) {
	buff := bytes.Buffer{}

	if err := encoding.EncodeNumber(&buff, g.GameId); err != nil {
		return nil, err
	}

	if err := encoding.EncodeNumber(&buff, g.AveragePlaytime); err != nil {
		return nil, err
	}

	if err := encoding.EncodeString(&buff, g.Name); err != nil {
		return nil, err
	}

	if err := encoding.EncodeString(&buff, g.Genres); err != nil {
		return nil, err
	}

	if err := encoding.EncodeString(&buff, g.ReleaseDate); err != nil {
		return nil, err
	}

	if err := encoding.EncodeBool(&buff, g.Windows); err != nil {
		return nil, err
	}

	if err := encoding.EncodeBool(&buff, g.Mac); err != nil {
		return nil, err
	}

	if err := encoding.EncodeBool(&buff, g.Linux); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (g Game) ToGameNamesMessage(genreToFilter string) GameNames {
	var result GameNames

	for _, h := range g {
		genres := strings.Split(h.Genres, sep)
		for _, genre := range genres {
			if genre == genreToFilter {
				result = append(result, GameName{GameId: h.GameId, GameName: h.Name})
				break
			}
		}
	}

	return result
}

func (g Game) ToGameReleasesMessage(genreToFilter string) Releases {
	var result Releases

	for _, h := range g {
		genres := strings.Split(h.Genres, sep)
		for _, genre := range genres {
			if genre == genreToFilter {
				result = append(result, release{GameId: h.GameId, GameName: h.Name, ReleaseDate: h.ReleaseDate, AvgPlaytime: h.AveragePlaytime})
				break
			}
		}
	}

	return result
}

func (g Game) ToPlatformMessage() Platform {
	var result Platform

	for _, h := range g {
		if h.Windows {
			result.Windows += 1
		}
		if h.Mac {
			result.Mac += 1
		}
		if h.Linux {
			result.Linux += 1
		}
	}

	return result
}
