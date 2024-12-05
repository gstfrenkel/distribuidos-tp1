package id

import (
	"os"
	"testing"
)

func TestIdGeneratorReturnsCorrectIdFormat(t *testing.T) {
	g := NewGenerator(1, "test-%d.csv")
	expected := "1-0"
	if id := g.GetId(); id != expected {
		t.Errorf("Expected id %s, got %s", expected, id)
	}
	g.Close()
	resetFiles()
}

func TestIdGeneratorIncrementsId(t *testing.T) {
	g := NewGenerator(0, "test-%d.csv")
	g.GetId() // 0-0
	expected := "0-1"
	if id := g.GetId(); id != expected {
		t.Errorf("Expected id %s, got %s", expected, id)
	}
	g.Close()
	resetFiles()
}

func TestPersistentIdGeneratorIncrementsId(t *testing.T) {
	g := NewGenerator(0, "test-%d.csv")
	g.GetId() // 0-0
	expected := "0-1"
	if id := g.GetId(); id != expected {
		t.Errorf("Expected id %s, got %s", expected, id)
	}
	g.Close()

	g2 := NewGenerator(0, "test-%d.csv")
	g2.GetId() // 0-0
	expected2 := "0-2"
	if id := g.GetId(); id != expected2 {
		t.Errorf("Expected id %s, got %s", expected2, id)
	}
	g.Close()
	resetFiles()
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

func resetFiles() {
	_ = os.Remove("test-0.csv")
	_ = os.Remove("test-1.csv")
}
