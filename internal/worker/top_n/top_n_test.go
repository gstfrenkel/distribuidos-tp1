package top_n

import (
	"container/heap"
	"testing"
	"tp1/pkg/message"
)

func TestPublishNewGame(t *testing.T) {
	f := fakeFilter(5)

	msg := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}
	f.updateTop(msg)
	if f.top.Len() != 1 {
		t.Errorf("Expected top length 1, got %d", f.top.Len())
	}
	if f.top[0].GameId != 1 || f.top[0].Votes != 10 {
		t.Errorf("Expected GameId 1 with 10 Votes, got GameId %d with %d Votes", f.top[0].GameId, f.top[0].Votes)
	}
}

func TestPublishExistingGame(t *testing.T) {
	f := fakeFilter(5)
	msg1 := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}
	msg2 := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 15}}
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
	f := fakeFilter(5)
	for i := 1; i <= 6; i++ {
		msg := message.ScoredReviews{message.ScoredReview{GameId: int64(i), Votes: uint64(i * 10)}}
		f.updateTop(msg)
	}
	if f.top.Len() != 5 {
		t.Errorf("Expected top length 5, got %d", f.top.Len())
	}
	if heap.Pop(&f.top).(*message.ScoredReview).Votes != 20 {
		t.Errorf("Expected min top vote 20, got %d", heap.Pop(&f.top).(*message.ScoredReview).Votes)
	}
}

func TestInsertTwoElementsWithSameVotes(t *testing.T) {
	f := fakeFilter(5)

	msg1 := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}
	msg2 := message.ScoredReviews{message.ScoredReview{GameId: 2, Votes: 10}}

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

func TestTopToTopNSlice(t *testing.T) {
	f := fakeFilter(2)
	msg := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}
	f.updateTop(msg)
	msg = message.ScoredReviews{message.ScoredReview{GameId: 2, Votes: 20}}
	f.updateTop(msg)
	msg = message.ScoredReviews{message.ScoredReview{GameId: 3, Votes: 30}}
	f.updateTop(msg)

	top := f.getTopNScoredReviews()
	if len(top) != 2 {
		t.Errorf("Expected top length 2, got %d", len(top))
	}
	if top[0].GameId != 3 || top[0].Votes != 30 {
		t.Errorf("Expected top GameId 3 with 30 Votes, got GameId %d with %d Votes", top[0].GameId, top[0].Votes)
	}
}

func fakeFilter(n int) *filter {
	return &filter{top: make(PriorityQueue, 0, n), n: n}
}
