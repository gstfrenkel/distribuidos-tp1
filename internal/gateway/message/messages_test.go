package message

import (
	"testing"
)

func TestToBytes_ValidReviewMsg(t *testing.T) {
	review := &ReviewMsg{
		appId:       123,
		appName:     "TestApp",
		reviewText:  "Great app!",
		reviewScore: 5,
		reviewVotes: 100,
	}

	data, err := review.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("expected non-empty byte slice")
	}
}

func TestFromBytes_ValidData(t *testing.T) {
	review := &ReviewMsg{
		appId:       123,
		appName:     "TestApp",
		reviewText:  "Great app!",
		reviewScore: 5,
		reviewVotes: 100,
	}

	data, err := review.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	newReview, err := review.FromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if newReview.appId != review.appId ||
		newReview.appName != review.appName ||
		newReview.reviewText != review.reviewText ||
		newReview.reviewScore != review.reviewScore ||
		newReview.reviewVotes != review.reviewVotes {
		t.Fatalf("expected %v, got %v", review, newReview)
	}
}

func TestFromBytes_InvalidData(t *testing.T) {
	invalidData := []byte{0, 1, 2, 3, 4}

	_, err := (&ReviewMsg{}).FromBytes(invalidData)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestToBytes_EmptyFields(t *testing.T) {
	review := &ReviewMsg{}

	data, err := review.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("expected non-empty byte slice")
	}
}

func TestFromBytes_EmptyFields(t *testing.T) {
	review := &ReviewMsg{}

	data, err := review.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	newReview, err := review.FromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if newReview.appId != review.appId ||
		newReview.appName != review.appName ||
		newReview.reviewText != review.reviewText ||
		newReview.reviewScore != review.reviewScore ||
		newReview.reviewVotes != review.reviewVotes {
		t.Fatalf("expected %v, got %v", review, newReview)
	}
}
