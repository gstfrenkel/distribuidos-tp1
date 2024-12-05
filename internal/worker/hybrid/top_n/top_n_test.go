package top_n

import (
	"container/heap"
	"testing"
	"tp1/pkg/message"
)

func TestPublishNewGame(t *testing.T) {
	f := fakeFilter(5)
	clientId := "0-0"
	msg, _ := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}.ToBytes()
	f.updateTop(msg, clientId)
	priorityQueue := f.top[clientId]
	if priorityQueue.Len() != 1 {
		t.Errorf("Expected top length 1, got %d", priorityQueue.Len())
	}
	if priorityQueue[0].GameId != 1 || priorityQueue[0].Votes != 10 {
		t.Errorf("Expected GameId 1 with 10 Votes, got GameId %d with %d Votes", priorityQueue[0].GameId, priorityQueue[0].Votes)
	}
}

func TestPublishExistingGame(t *testing.T) {
	f := fakeFilter(5)
	clientId := "0-0"
	msg1, _ := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}.ToBytes()
	msg2, _ := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 15}}.ToBytes()
	f.updateTop(msg1, clientId)
	f.updateTop(msg2, clientId)
	priorityQueue := f.top[clientId]
	if priorityQueue.Len() != 1 {
		t.Errorf("Expected top length 1, got %d", priorityQueue.Len())
	}
	if priorityQueue[0].GameId != 1 || priorityQueue[0].Votes != 15 {
		t.Errorf("Expected GameId 1 with 15 Votes, got GameId %d with %d Votes", priorityQueue[0].GameId, priorityQueue[0].Votes)
	}
}

func TestPublishTopNLimit(t *testing.T) {
	f := fakeFilter(5)
	clientId := "0-0"
	for i := 1; i <= 6; i++ {
		msg, _ := message.ScoredReviews{message.ScoredReview{GameId: int64(i), Votes: uint64(i * 10)}}.ToBytes()
		f.updateTop(msg, clientId)
	}
	priorityQueue := f.top[clientId]
	if priorityQueue.Len() != 5 {
		t.Errorf("Expected top length 5, got %d", priorityQueue.Len())
	}
	if heap.Pop(&priorityQueue).(*message.ScoredReview).Votes != 20 {
		t.Errorf("Expected min top vote 20, got %d", heap.Pop(&priorityQueue).(*message.ScoredReview).Votes)
	}
}

func TestInsertTwoElementsWithSameVotes(t *testing.T) {
	f := fakeFilter(5)
	clientId := "0-0"
	msg1, _ := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}.ToBytes()
	msg2, _ := message.ScoredReviews{message.ScoredReview{GameId: 2, Votes: 10}}.ToBytes()

	f.updateTop(msg1, clientId)
	f.updateTop(msg2, clientId)

	priorityQueue := f.top[clientId]
	if priorityQueue.Len() != 2 {
		t.Errorf("Expected top length 2, got %d", priorityQueue.Len())
	}

	foundGameId1 := false
	foundGameId2 := false

	for _, item := range priorityQueue {
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
	clientId := "0-0"
	msg, _ := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}.ToBytes()
	f.updateTop(msg, clientId)
	msg, _ = message.ScoredReviews{message.ScoredReview{GameId: 2, Votes: 20}}.ToBytes()
	f.updateTop(msg, clientId)
	msg, _ = message.ScoredReviews{message.ScoredReview{GameId: 3, Votes: 30}}.ToBytes()
	f.updateTop(msg, clientId)

	top := f.getTopNScoredReviews(clientId)
	if len(top) != 2 {
		t.Errorf("Expected top length 2, got %d", len(top))
	}
	if top[0].GameId != 3 || top[0].Votes != 30 {
		t.Errorf("Expected top GameId 3 with 30 Votes, got GameId %d with %d Votes", top[0].GameId, top[0].Votes)
	}
}

func TestManyClientTops(t *testing.T) {
	f := fakeFilter(2)
	clientId1 := "0-0"
	clientId2 := "0-1"
	msg, _ := message.ScoredReviews{message.ScoredReview{GameId: 1, Votes: 10}}.ToBytes()
	f.updateTop(msg, clientId1)
	msg, _ = message.ScoredReviews{message.ScoredReview{GameId: 2, Votes: 20}}.ToBytes()
	f.updateTop(msg, clientId1)
	msg, _ = message.ScoredReviews{message.ScoredReview{GameId: 3, Votes: 30}}.ToBytes()
	f.updateTop(msg, clientId1)

	msg, _ = message.ScoredReviews{message.ScoredReview{GameId: 4, Votes: 40}}.ToBytes()
	f.updateTop(msg, clientId2)
	msg, _ = message.ScoredReviews{message.ScoredReview{GameId: 5, Votes: 50}}.ToBytes()
	f.updateTop(msg, clientId2)
	msg, _ = message.ScoredReviews{message.ScoredReview{GameId: 6, Votes: 60}}.ToBytes()
	f.updateTop(msg, clientId2)

	top1 := f.getTopNScoredReviews(clientId1)
	if len(top1) != 2 {
		t.Errorf("Expected top1 length 2, got %d", len(top1))
	}
	if top1[0].GameId != 3 || top1[0].Votes != 30 {
		t.Errorf("Expected top1 GameId 3 with 30 Votes, got GameId %d with %d Votes", top1[0].GameId, top1[0].Votes)
	}

	top2 := f.getTopNScoredReviews(clientId2)
	if len(top2) != 2 {
		t.Errorf("Expected top2 length 2, got %d", len(top2))
	}
	if top2[0].GameId != 6 || top2[0].Votes != 60 {
		t.Errorf("Expected top2 GameId 6 with 60 Votes, got GameId %d with %d Votes", top2[0].GameId, top2[0].Votes)
	}
}

func fakeFilter(n int) *filter {
	return &filter{top: make(map[string]PriorityQueue), n: n}
}
