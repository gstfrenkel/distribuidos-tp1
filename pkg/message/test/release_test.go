package test_test

import (
	"testing"

	"tp1/pkg/message"

	"github.com/stretchr/testify/assert"
)

func Test_Release(t *testing.T) {
	exp := message.Releases{
		{GameId: 123, GameName: "AWasdoawijpdajdsaopjdjw", ReleaseDate: "123109238123897128973128793", AvgPlaytime: 1283991},
		{GameId: 56464, GameName: "sdNOWIdwajidwaopijdawopidjpo", ReleaseDate: "2525345346376474567", AvgPlaytime: 807097},
		{GameId: 12, GameName: "AWasdoawijpdajdsaopjdjw", ReleaseDate: "356745754747", AvgPlaytime: 4567456756},
		{GameId: 8786, GameName: "AWjaf89jpaw89fd89a0wf8y9", ReleaseDate: "532534538490580934", AvgPlaytime: 342524},
		{GameId: 4, GameName: "AOPiefjdiajfapoidjwpadjpiapoiwdja", ReleaseDate: "2323234", AvgPlaytime: 6574},
	}

	b, err := exp.ToBytes()
	assert.NoError(t, err)

	res, err := message.ReleasesFromBytes(b)
	assert.NoError(t, err)

	assert.Equal(t, exp, res)
}
