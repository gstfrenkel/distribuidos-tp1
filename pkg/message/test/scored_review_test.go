package test_test

import (
	"testing"

	"tp1/pkg/message"

	"github.com/stretchr/testify/assert"
)

func TestScoredReviewToBytes(t *testing.T) {
	reviews := message.ScoredReviews{
		{GameId: 1, GameName: "Game1", Votes: 4},
		{GameId: 2, GameName: "Game1", Votes: 1},
		{GameId: 3, GameName: "Game1", Votes: 8},
	}

	for _, review := range reviews {
		b, err := review.ToBytes()
		assert.NoError(t, err)
		assert.NotEmpty(t, b)
	}
}

// Test for ScoredReviewFromBytes method
func TestScoredReviewFromBytes(t *testing.T) {
	reviews := message.ScoredReviews{
		{GameId: 1, GameName: "Game1", Votes: 4},
		{GameId: 2, GameName: "Game1", Votes: 1},
		{GameId: 3, GameName: "Game1", Votes: 8},
	}

	for i, review := range reviews {
		b, err := review.ToBytes()
		assert.NoError(t, err)

		deserialized, err := message.ScoredReviewFromBytes(b)
		assert.NoError(t, err)

		assert.Equal(t, reviews[i], deserialized)

	}
}
