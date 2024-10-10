package test_test

import (
	"testing"

	"tp1/pkg/message"

	"github.com/stretchr/testify/assert"
)

func TestReviewMessageToBytes(t *testing.T) {
	original := message.Review{
		{GameId: 1, GameName: "Game1", Text: "Great action", Score: 1},
		{GameId: 1, GameName: "Game1", Text: "Bad action", Score: -1},
		{GameId: 2, GameName: "Game2", Text: "Bad action", Score: -1},
	}

	serialized, err := original.ToBytes()

	assert.NoError(t, err)
	assert.NotEmpty(t, serialized)
}

func TestReviewMessageFromBytes(t *testing.T) {
	original := message.Review{
		{GameId: 1, GameName: "Game1", Text: "Great action", Score: 1},
		{GameId: 1, GameName: "Game1", Text: "Bad action", Score: -1},
		{GameId: 2, GameName: "Game2", Text: "Bad action", Score: -1},
	}

	serialized, err := original.ToBytes()
	assert.NoError(t, err)

	deserialized, err := message.ReviewFromBytes(serialized)
	assert.NoError(t, err)

	assert.Equal(t, original, deserialized)
}
