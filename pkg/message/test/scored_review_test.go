package test_test

import (
	"testing"

	"tp1/pkg/message"

	"github.com/stretchr/testify/assert"
)

func Test_ScoredReview(t *testing.T) {
	reviews := message.ScoredReviews{
		{GameId: 134, GameName: "Game1a8awdp8aw9d8a9wjd8a9jd83q1dj01", Votes: 544},
		{GameId: 5462, GameName: "Game1adawoud8a9jdj91d", Votes: 15464},
		{GameId: 3756, GameName: "Game1", Votes: 2452},
	}

	for i, review := range reviews {
		b, err := review.ToBytes()
		assert.NoError(t, err)

		res, err := message.ScoredReviewFromBytes(b)
		assert.NoError(t, err)

		assert.Equal(t, reviews[i], res)

	}
}

func Test_ScoredReviews(t *testing.T) {
	reviews := message.ScoredReviews{
		{GameId: 134, GameName: "Game1a8awdp8aw9d8a9wjd8a9jd83q1dj01", Votes: 544},
		{GameId: 5462, GameName: "Game1adawoud8a9jdj91d", Votes: 15464},
		{GameId: 3756, GameName: "Game1", Votes: 2452},
	}

	b, err := reviews.ToBytes()
	assert.NoError(t, err)

	res, err := message.ScoredReviewsFromBytes(b)
	assert.NoError(t, err)

	assert.Equal(t, reviews, res)
}
