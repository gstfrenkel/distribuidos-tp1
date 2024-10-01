package message

type Game []game

type game struct {
	GameId          int64
	Name            string
	Windows         bool
	Mac             bool
	Linux           bool
	Genres          string
	AveragePlaytime int64
	ReleaseDate     string
}

func (g Game) ToBytes() ([]byte, error) {
	return toBytes(g)
}

func GameFromClientGame(clientGame []DataCSVGames) ([]byte, error) {
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
