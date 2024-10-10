package test

import (
	"testing"

	"tp1/pkg/message"

	"github.com/stretchr/testify/assert"
)

func TestEof_ToBytes_ValidInput(t *testing.T) {
	eof := message.Eof{1, 2, 3}
	expected := []byte{3, 1, 2, 3}

	result, err := eof.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, expected, result)
}

func TestEof_ToBytes_EmptyInput(t *testing.T) {
	eof := message.Eof{}
	expected := []byte{0}

	result, err := eof.ToBytes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, expected, result)
}

func TestEofFromBytes_ValidInput(t *testing.T) {
	input := []byte{3, 1, 2, 3}
	expected := message.Eof{1, 2, 3}

	result, err := message.EofFromBytes(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, expected, result)
}

func TestEofFromBytes_EmptyInput(t *testing.T) {
	input := []byte{0}
	expected := message.Eof{}

	result, err := message.EofFromBytes(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, expected, result)
}

func TestEof_Contains_True(t *testing.T) {
	eof := message.Eof{1, 2, 3}
	assert.True(t, eof.Contains(2))
}

func TestEof_Contains_False(t *testing.T) {
	eof := message.Eof{1, 2, 3}
	assert.False(t, eof.Contains(4))
}
