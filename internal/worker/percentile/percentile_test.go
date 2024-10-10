package percentile

import (
	"testing"
	"tp1/pkg/message"
)

func TestSaveScoredReviewAppendsMessages(t *testing.T) {
	f := &filter{}
	msg := message.ScoredReviews{
		{GameId: 1, Votes: 10},
		{GameId: 2, Votes: 20},
	}

	f.saveScoredReview(msg)

	if len(f.scoredReviews) != 2 {
		t.Errorf("Expected 2 scored reviews, got %d", len(f.scoredReviews))
	}
}

func TestGetGamesInPercentileReturnsCorrectSubset(t *testing.T) {
	f := &filter{n: 90}
	f.scoredReviews = message.ScoredReviews{
		{GameId: 1, Votes: 10},
		{GameId: 2, Votes: 20},
		{GameId: 3, Votes: 30},
		{GameId: 4, Votes: 40},
		{GameId: 5, Votes: 50},
	}

	result := f.getGamesInPercentile()

	expectedLength := 1
	if len(result) != expectedLength {
		t.Errorf("Expected %d scored reviews, got %d", expectedLength, len(result))
	}

	expectedGameId := int64(5)
	if result[0].GameId != expectedGameId {
		t.Errorf("Expected GameId %d, got %d", expectedGameId, result[0].GameId)
	}
}

func TestPercentileIdxCalculatesCorrectIndex(t *testing.T) {
	f := &filter{n: 90}
	f.scoredReviews = message.ScoredReviews{
		{GameId: 1, Votes: 10},
		{GameId: 2, Votes: 20},
		{GameId: 3, Votes: 30},
		{GameId: 4, Votes: 40},
		{GameId: 5, Votes: 50},
	}

	idx := f.percentileIdx()

	expectedIdx := 4
	if idx != expectedIdx {
		t.Errorf("Expected index %d, got %d", expectedIdx, idx)
	}
}

func TestSortScoredReviewsSortsCorrectly(t *testing.T) {
	f := &filter{}
	f.scoredReviews = message.ScoredReviews{
		{GameId: 3, Votes: 30},
		{GameId: 1, Votes: 10},
		{GameId: 2, Votes: 20},
	}

	f.sortScoredReviews()

	if f.scoredReviews[0].Votes != 10 || f.scoredReviews[1].Votes != 20 || f.scoredReviews[2].Votes != 30 {
		t.Errorf("Scored reviews not sorted correctly")
	}
}
