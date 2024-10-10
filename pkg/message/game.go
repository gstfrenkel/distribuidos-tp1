package message

import (
	"strings"
)

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

func GameFromBytes(b []byte) (Game, error) {
	var m Game
	return m, fromBytes(b, &m)
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
	return toBytes(g)
}

func (g Game) ToGameNamesMessage(genreToFilter string) GameNames {
	var result GameNames

	for _, h := range g {
		genres := strings.Split(h.Genres, ",")
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
		genres := strings.Split(h.Genres, ",")
		for _, genre := range genres {
			if genre == genreToFilter {
				result = append(result, Release{GameId: h.GameId, GameName: h.Name, ReleaseDate: h.ReleaseDate, AvgPlaytime: h.AveragePlaytime})
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
