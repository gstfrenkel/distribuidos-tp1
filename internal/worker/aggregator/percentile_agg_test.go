package aggregator

import (
	"testing"

	"tp1/pkg/message"
)

func TestSaveScoredReviewAppendsMessages(t *testing.T) {
	f := fakeFilter(90)
	msg := message.ScoredReviews{
		{GameId: 1, Votes: 10},
		{GameId: 2, Votes: 20},
	}

	clientId := "0-0"
	bytes, _ := msg.ToBytes()
	f.save(bytes, clientId)

	if len(f.scoredReviews[clientId]) != 2 {
		t.Errorf("Expected 2 scored reviews, got %d", len(f.scoredReviews[clientId]))
	}

	if len(f.scoredReviews) != 1 {
		t.Errorf("Expected map len 1, got %d", len(f.scoredReviews))
	}
}

func TestGetGamesInPercentileReturnsCorrectSubset(t *testing.T) {
	f := fakeFilter(90)
	clientId := "0-0"
	f.scoredReviews = map[string]message.ScoredReviews{
		clientId: {
			{GameId: 1, Votes: 10},
			{GameId: 2, Votes: 20},
			{GameId: 3, Votes: 30},
			{GameId: 4, Votes: 40},
			{GameId: 5, Votes: 50},
		},
	}

	result := f.getGamesInPercentile(clientId)

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
	f := fakeFilter(90)
	clientId := "0-0"
	f.scoredReviews = map[string]message.ScoredReviews{
		clientId: {
			{GameId: 1, Votes: 10},
			{GameId: 2, Votes: 20},
			{GameId: 3, Votes: 30},
			{GameId: 4, Votes: 40},
			{GameId: 5, Votes: 50},
		},
	}

	idx := f.percentileIdx(clientId)

	expectedIdx := 4
	if idx != expectedIdx {
		t.Errorf("Expected index %d, got %d", expectedIdx, idx)
	}
}

func TestSortScoredReviewsSortsCorrectly(t *testing.T) {
	f := fakeFilter(90)
	clientId := "0-0"
	f.scoredReviews = map[string]message.ScoredReviews{
		clientId: {
			{GameId: 3, Votes: 30},
			{GameId: 1, Votes: 10},
			{GameId: 2, Votes: 20},
		},
	}

	f.scoredReviews[clientId].Sort(true)

	if f.scoredReviews[clientId][0].Votes != 10 || f.scoredReviews[clientId][1].Votes != 20 || f.scoredReviews[clientId][2].Votes != 30 {
		t.Errorf("Scored reviews not sorted correctly")
	}
}

func TestSaveScoredReviewAppendsMessagesToManyClients(t *testing.T) {
	f := fakeFilter(90)
	msg := message.ScoredReviews{
		{GameId: 1, Votes: 10},
		{GameId: 2, Votes: 20},
	}

	bytes, _ := msg.ToBytes()

	f.save(bytes, "0-0")
	f.save(bytes, "0-1")

	if len(f.scoredReviews) != 2 {
		t.Errorf("Expected map len 2, got %d", len(f.scoredReviews))
	}

	if len(f.scoredReviews["0-0"]) != 2 {
		t.Errorf("Expected 2 scored reviews for client 0-0, got %d", len(f.scoredReviews["0-0"]))
	}

	if len(f.scoredReviews["0-1"]) != 2 {
		t.Errorf("Expected 2 scored reviews for client 0-1, got %d", len(f.scoredReviews["0-1"]))
	}
}

func fakeFilter(n uint8) *percentile {
	return &percentile{n: n, scoredReviews: make(map[string]message.ScoredReviews)}
}
