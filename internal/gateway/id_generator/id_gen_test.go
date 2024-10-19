package id_generator

import (
	"testing"
)

func TestIdGeneratorReturnsCorrectIdFormat(t *testing.T) {
	g := New(1)
	expected := "1-0"
	if id := g.GetId(); id != expected {
		t.Errorf("Expected id %s, got %s", expected, id)
	}
}

func TestIdGeneratorIncrementsId(t *testing.T) {
	g := New(1)
	g.GetId() // 1-0
	expected := "1-1"
	if id := g.GetId(); id != expected {
		t.Errorf("Expected id %s, got %s", expected, id)
	}
}
