package encoding

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

func TestEncodeClientId(t *testing.T) {
	clientId := "0-1"
	expectedLength := 32
	expectedEncoded := make([]byte, expectedLength)
	copy(expectedEncoded, clientId)

	encoded := EncodeClientId(clientId)

	if len(encoded) != expectedLength {
		t.Errorf("Expected encoded length %d, got %d", expectedLength, len(encoded))
	}

	for i := 0; i < expectedLength; i++ {
		if encoded[i] != expectedEncoded[i] {
			t.Errorf("Expected byte at position %d to be %d, got %d", i, expectedEncoded[i], encoded[i])
		}
	}
}

func TestEncodeClientIdEndingWith0(t *testing.T) {
	clientId := "0-10"
	expectedLength := 32
	expectedEncoded := make([]byte, expectedLength)
	copy(expectedEncoded, clientId)

	encoded := EncodeClientId(clientId)

	if len(encoded) != expectedLength {
		t.Errorf("Expected encoded length %d, got %d", expectedLength, len(encoded))
	}

	for i := 0; i < expectedLength; i++ {
		if encoded[i] != expectedEncoded[i] {
			t.Errorf("Expected byte at position %d to be %d, got %d", i, expectedEncoded[i], encoded[i])
		}
	}
}

func TestDecodeClientId(t *testing.T) {
	clientId := "0-1"
	encoded := EncodeClientId(clientId)
	decoded := DecodeClientId(encoded)

	if decoded != clientId {
		t.Errorf("Expected decoded client id %s, got %s", clientId, decoded)
	}
}

func TestDecodeClientIdEndingWith0(t *testing.T) {
	clientId := "0-10"
	encoded := EncodeClientId(clientId)
	decoded := DecodeClientId(encoded)

	if decoded != clientId {
		t.Errorf("Expected decoded client id %s, got %s", clientId, decoded)
	}
}
