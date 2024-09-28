package message

import (
	"testing"
	"tp1/pkg/message/review"
)

func TestToBytes_ValidReviewMsg(t *testing.T) {
	r := review.New(123, "TestApp", "Great app!", 5, 100)

	data, err := r.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("expected non-empty byte slice")
	}
}

func TestFromBytes_ValidData(t *testing.T) {
	r := review.New(123, "TestApp", "Great app!", 5, 100)

	data, err := r.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	newReview, err := review.FromBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	b1, _ := newReview.ToBytes()
	b2, _ := r.ToBytes()

	if string(b1) != string(b2) {
		t.Fatalf("expected %v, got %v", r, newReview)
	}
}

func TestFromBytes_InvalidData(t *testing.T) {
	invalidData := []byte{0, 1, 2, 3, 4}

	_, err := review.FromBytes()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestFromBytes_EmptyFields(t *testing.T) {
	r := review.New(0, "", "", 0, 0)

	data, err := r.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	newReview, err := review.FromBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	b1, _ := newReview.ToBytes()
	b2, _ := r.ToBytes()

	if string(b1) != string(b2) {
		t.Fatalf("expected %v, got %v", r, newReview)
	}
}
