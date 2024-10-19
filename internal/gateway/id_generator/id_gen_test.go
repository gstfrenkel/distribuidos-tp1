package id_generator

import (
	"bytes"
	"encoding/gob"
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

func TestIdGeneratorEncodesCorrectly(t *testing.T) {
	clientId := "0-1"
	encoded := EncodeClientId(clientId)

	var decodedClientId string
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&decodedClientId)
	if err != nil {
		t.Errorf("Error decoding client id: %s", err)
	}

	if decodedClientId != clientId {
		t.Errorf("Expected decoded client id %s, got %s", clientId, decodedClientId)
	}
}
