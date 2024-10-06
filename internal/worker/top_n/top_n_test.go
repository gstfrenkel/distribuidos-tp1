package top_n

import (
	"container/heap"
	"testing"
	"tp1/pkg/message"
)

func TestUpdateVotes(t *testing.T) {
	f := &filter{votes: make(map[GameId]uint64)}
	f.updateVotes(1, 10)
	if f.votes[1] != 10 {
		t.Errorf("Expected 10 Votes, got %d", f.votes[1])
	}
	f.updateVotes(1, 5)
	if f.votes[1] != 15 {
		t.Errorf("Expected 15 Votes, got %d", f.votes[1])
	}
}

func TestPublishNewGame(t *testing.T) {
	f := fakeFilter()

	msg := message.ScoredReview{GameId: 1, Votes: 10}
	f.updateTop(msg)
	if f.top.Len() != 1 {
		t.Errorf("Expected top length 1, got %d", f.top.Len())
	}
	if f.top[0].GameId != 1 || f.top[0].Votes != 10 {
		t.Errorf("Expected GameId 1 with 10 Votes, got GameId %d with %d Votes", f.top[0].GameId, f.top[0].Votes)
	}
}

func TestPublishExistingGame(t *testing.T) {
	f := fakeFilter()
	msg1 := message.ScoredReview{GameId: 1, Votes: 10}
	msg2 := message.ScoredReview{GameId: 1, Votes: 5}
	f.updateTop(msg1)
	f.updateTop(msg2)
	if f.top.Len() != 1 {
		t.Errorf("Expected top length 1, got %d", f.top.Len())
	}
	if f.top[0].GameId != 1 || f.top[0].Votes != 15 {
		t.Errorf("Expected GameId 1 with 15 Votes, got GameId %d with %d Votes", f.top[0].GameId, f.top[0].Votes)
	}
}

func TestPublishTopNLimit(t *testing.T) {
	f := fakeFilter()
	for i := 1; i <= 6; i++ {
		msg := message.ScoredReview{GameId: int64(i), Votes: uint64(i * 10)}
		f.updateTop(msg)
	}
	if f.top.Len() != 5 {
		t.Errorf("Expected top length 5, got %d", f.top.Len())
	}
	if heap.Pop(&f.top).(*message.ScoredReview).Votes != 60 {
		t.Errorf("Expected max top vote 60, got %d", heap.Pop(&f.top).(*message.ScoredReview).Votes)
	}
}

func TestInsertTwoElementsWithSameVotes(t *testing.T) {
	f := fakeFilter()

	msg1 := message.ScoredReview{GameId: 1, Votes: 10}
	msg2 := message.ScoredReview{GameId: 2, Votes: 10}

	f.updateTop(msg1)
	f.updateTop(msg2)

	if f.top.Len() != 2 {
		t.Errorf("Expected top length 2, got %d", f.top.Len())
	}

	foundGameId1 := false
	foundGameId2 := false

	for _, item := range f.top {
		if item.GameId == 1 && item.Votes == 10 {
			foundGameId1 = true
		}
		if item.GameId == 2 && item.Votes == 10 {
			foundGameId2 = true
		}
	}

	if !foundGameId1 {
		t.Errorf("Expected to find GameId 1 with 10 Votes")
	}
	if !foundGameId2 {
		t.Errorf("Expected to find GameId 2 with 10 Votes")
	}
}

func fakeFilter() *filter {
	return &filter{votes: make(map[GameId]uint64), top: make(PriorityQueue, 0, 5), n: 5}
}
