package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"tp1/pkg/message"
)

func TestGameEncodingDecoding(t *testing.T) {
	originalGames := message.Game{
		{
			GameId:          1,
			AveragePlaytime: 100,
			Name:            "Game A",
			Genres:          "Action,Adventure",
			ReleaseDate:     "2022-01-01",
			Windows:         true,
			Mac:             false,
			Linux:           true,
		},
		{
			GameId:          2,
			AveragePlaytime: 200,
			Name:            "Game B",
			Genres:          "RPG,Strategy",
			ReleaseDate:     "2021-12-01",
			Windows:         false,
			Mac:             true,
			Linux:           true,
		},
	}

	encodedBytes, err := originalGames.ToBytes()
	assert.NoError(t, err, "Encoding should not produce an error")
	assert.NotNil(t, encodedBytes, "Encoded bytes should not be nil")

	decodedGames, err := message.GamesFromBytes(encodedBytes)
	assert.NoError(t, err, "Decoding should not produce an error")
	assert.NotNil(t, decodedGames, "Decoded games should not be nil")

	assert.Equal(t, originalGames, decodedGames, "Original and decoded games should be the same")
}

func TestEmptyGameList(t *testing.T) {
	emptyGames := message.Game{}

	encodedBytes, err := emptyGames.ToBytes()
	assert.NoError(t, err, "Encoding empty games should not produce an error")
	assert.NotNil(t, encodedBytes, "Encoded bytes for empty games should not be nil")

	decodedGames, err := message.GamesFromBytes(encodedBytes)
	assert.NoError(t, err, "Decoding empty games should not produce an error")
	assert.Empty(t, decodedGames, "Decoded games should be empty")
}
