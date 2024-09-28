package review

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToBytes(t *testing.T) {
	original := Message{
		{GameId: 1, GameName: "Game1", Text: "Great game", Score: 1},
		{GameId: 1, GameName: "Game1", Text: "Bad game", Score: -1},
		{GameId: 2, GameName: "Game2", Text: "Bad game", Score: -1},
	}

	serialized, err := original.ToBytes()

	assert.NoError(t, err)
	assert.NotEmpty(t, serialized)
}

// Test for FromBytes method
func TestFromBytes(t *testing.T) {
	original := Message{
		{GameId: 1, GameName: "Game1", Text: "Great game", Score: 1},
		{GameId: 1, GameName: "Game1", Text: "Bad game", Score: -1},
		{GameId: 2, GameName: "Game2", Text: "Bad game", Score: -1},
	}

	serialized, err := original.ToBytes()
	assert.NoError(t, err)

	deserialized, err := FromBytes(serialized)
	assert.NoError(t, err)

	assert.Equal(t, original, deserialized)
}
