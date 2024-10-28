package percentile

import (
	"errors"
	"reflect"
	"testing"
	"tp1/pkg/message"
)

func TestNextBatchReturnsCorrectBatch(t *testing.T) {
	games := []any{
		message.ScoredReview{GameId: 1, Votes: 10},
		message.ScoredReview{GameId: 2, Votes: 20},
		message.ScoredReview{GameId: 3, Votes: 30},
		message.ScoredReview{GameId: 4, Votes: 40},
	}

	batch, nextStart := nextBatch(games, len(games), 2, 0)
	expectedBatch := []any{
		message.ScoredReview{GameId: 1, Votes: 10},
		message.ScoredReview{GameId: 2, Votes: 20},
	}

	if !reflect.DeepEqual(batch, expectedBatch) {
		t.Errorf("Expected batch %v, got %v", expectedBatch, batch)
	}

	if nextStart != 2 {
		t.Errorf("Expected next start index 2, got %d", nextStart)
	}
}

func TestNextBatchHandlesEndOfSlice(t *testing.T) {
	games := []any{
		message.ScoredReview{GameId: 1, Votes: 10},
		message.ScoredReview{GameId: 2, Votes: 20},
	}

	batch, nextStart := nextBatch(games, len(games), 3, 0)
	expectedBatch := []any{
		message.ScoredReview{GameId: 1, Votes: 10},
		message.ScoredReview{GameId: 2, Votes: 20},
	}

	if !reflect.DeepEqual(batch, expectedBatch) {
		t.Errorf("Expected batch %v, got %v", expectedBatch, batch)
	}

	if nextStart != 2 {
		t.Errorf("Expected next start index 2, got %d", nextStart)
	}
}

func TestNextBatchHandlesEmptySlice(t *testing.T) {
	var games []any

	batch, nextStart := nextBatch(games, len(games), 2, 0)
	var expectedBatch []any

	if !reflect.DeepEqual(batch, expectedBatch) {
		t.Errorf("Expected batch %v, got %v", expectedBatch, batch)
	}

	if nextStart != 0 {
		t.Errorf("Expected next start index 0, got %d", nextStart)
	}
}

func TestSendBatchesSendsCorrectBatches(t *testing.T) {
	data := []any{
		message.ScoredReview{GameId: 1, Votes: 10},
		message.ScoredReview{GameId: 2, Votes: 20},
		message.ScoredReview{GameId: 3, Votes: 30},
		message.ScoredReview{GameId: 4, Votes: 40},
	}

	var sentBatches [][]byte
	toBytes := func(batch []any) ([]byte, error) {
		return []byte("batch"), nil
	}
	sendBatch := func(bytes []byte, _ string) {
		sentBatches = append(sentBatches, bytes)
	}

	SendBatches(data, 2, toBytes, sendBatch, "0-0")

	expectedBatches := [][]byte{[]byte("batch"), []byte("batch")}
	if !reflect.DeepEqual(sentBatches, expectedBatches) {
		t.Errorf("Expected batches %v, got %v", expectedBatches, sentBatches)
	}
}

func TestSendBatchesSendsScoredRevCorrectBatches(t *testing.T) {
	data := message.ScoredReviews{
		message.ScoredReview{GameId: 1, Votes: 10},
		message.ScoredReview{GameId: 2, Votes: 20},
		message.ScoredReview{GameId: 3, Votes: 30},
		message.ScoredReview{GameId: 4, Votes: 40},
	}.ToAny()

	var sentBatches [][]byte
	toBytes := func(batch []any) ([]byte, error) {
		return []byte("batch"), nil
	}
	sendBatch := func(bytes []byte, _ string) {
		sentBatches = append(sentBatches, bytes)
	}

	SendBatches(data, 2, toBytes, sendBatch, "0-0")

	expectedBatches := [][]byte{[]byte("batch"), []byte("batch")}
	if !reflect.DeepEqual(sentBatches, expectedBatches) {
		t.Errorf("Expected batches %v, got %v", expectedBatches, sentBatches)
	}
}

func TestSendBatchesHandlesEmptyData(t *testing.T) {
	var data []any

	var sentBatches [][]byte
	toBytes := func(batch []any) ([]byte, error) {
		return []byte("batch"), nil
	}
	sendBatch := func(bytes []byte, _ string) {
		sentBatches = append(sentBatches, bytes)
	}

	SendBatches(data, 2, toBytes, sendBatch, "0-0")

	if len(sentBatches) != 0 {
		t.Errorf("Expected no batches, got %v", sentBatches)
	}
}

func TestSendBatchesHandlesToBytesError(t *testing.T) {
	data := []any{
		message.ScoredReview{GameId: 1, Votes: 10},
		message.ScoredReview{GameId: 2, Votes: 20},
	}

	var sentBatches [][]byte
	toBytes := func(batch []any) ([]byte, error) {
		return nil, errors.New("toBytes error")
	}
	sendBatch := func(bytes []byte, _ string) {
		sentBatches = append(sentBatches, bytes)
	}

	SendBatches(data, 2, toBytes, sendBatch, "0-0")

	if len(sentBatches) != 0 {
		t.Errorf("Expected no batches due to error, got %v", sentBatches)
	}
}
