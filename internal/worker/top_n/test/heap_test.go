package test

import (
	"container/heap"
	"testing"
	"tp1/pkg/message"
)

func TestPriorityQueuePush(t *testing.T) {
	pq := &PriorityQueue{}

	item1 := &message.ScoredReview{GameId: 1, Votes: 10}
	item2 := &message.ScoredReview{GameId: 2, Votes: 20}
	item3 := &message.ScoredReview{GameId: 3, Votes: 5}

	heap.Push(pq, item1)
	heap.Push(pq, item2)
	heap.Push(pq, item3)

	if pq.Len() != 3 {
		t.Errorf("Expected length 3, got %d", pq.Len())
	}

	max1 := heap.Pop(pq).(*message.ScoredReview)
	max2 := heap.Pop(pq).(*message.ScoredReview)
	max3 := heap.Pop(pq).(*message.ScoredReview)

	if max1.Votes != 5 {
		t.Errorf("Expected max element with 5 votes, got %d", max1.Votes)
	}

	if max2.Votes != 10 {
		t.Errorf("Expected max element with 10 votes, got %d", max2.Votes)
	}

	if max3.Votes != 20 {
		t.Errorf("Expected max element with 20 votes, got %d", max3.Votes)
	}

	if pq.Len() != 0 {
		t.Errorf("Expected length 0, got %d", pq.Len())
	}
}

func TestPriorityQueueFix(t *testing.T) {
	pq := &PriorityQueue{}

	item1 := &message.ScoredReview{GameId: 1, Votes: 10}
	item2 := &message.ScoredReview{GameId: 2, Votes: 20}
	item3 := &message.ScoredReview{GameId: 3, Votes: 5}

	heap.Push(pq, item1)
	heap.Push(pq, item2)
	heap.Push(pq, item3)

	(*pq)[0].Votes = 6
	heap.Fix(pq, 0)

	if (*pq)[0].Votes != 6 {
		t.Errorf("Expected top element with 6 votes, got %d", (*pq)[0].Votes)
	} else {
		t.Log("When changing the top element to the lower value, the top element should be the same")
	}

	(*pq)[0].Votes = 60
	heap.Fix(pq, 0)

	if (*pq)[0].Votes != 10 {
		t.Errorf("Expected top element with 6 votes, got %d", (*pq)[0].Votes)
	} else {
		t.Log("When changing the top element to the higher value, the heap should re-arrange itself")
	}
}
