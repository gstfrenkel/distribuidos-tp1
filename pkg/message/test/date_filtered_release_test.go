package test

import (
	"bytes"
	"testing"

	"tp1/pkg/message"
)

func TestDateFilteredRelease_ToBytesAndFromBytes(t *testing.T) {
	original := message.DateFilteredRelease{
		GameId:      1,
		GameName:    "Test Game",
		AvgPlaytime: 120,
	}

	b, err := original.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	buffer := bytes.NewBuffer(b)
	decoded, err := message.DateFilteredReleaseFromBytes(buffer)
	if err != nil {
		t.Fatalf("DateFilteredReleaseFromBytes failed: %v", err)
	}

	if original != decoded {
		t.Fatalf("Expected %+v, got %+v", original, decoded)
	}
}

func TestDateFilteredReleases_ToBytesAndFromBytes(t *testing.T) {
	original := message.DateFilteredReleases{
		{GameId: 1, GameName: "Game A", AvgPlaytime: 100},
		{GameId: 2, GameName: "Game B", AvgPlaytime: 200},
		{GameId: 3, GameName: "Game C", AvgPlaytime: 300},
	}

	// Serialize to bytes
	b, err := original.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	// Deserialize back to struct
	decoded, err := message.DateFilteredReleasesFromBytes(b)
	if err != nil {
		t.Fatalf("DateFilteredReleasesFromBytes failed: %v", err)
	}

	// Assert equality
	if len(original) != len(decoded) {
		t.Fatalf("Expected length %d, got %d", len(original), len(decoded))
	}

	for i, originalRelease := range original {
		if originalRelease != decoded[i] {
			t.Errorf("At index %d, expected %+v, got %+v", i, originalRelease, decoded[i])
		}
	}
}

func TestDateFilteredRelease_EdgeCases(t *testing.T) {
	// Test with empty GameName
	release := message.DateFilteredRelease{
		GameId:      42,
		GameName:    "",
		AvgPlaytime: 0,
	}

	b, err := release.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	buffer := bytes.NewBuffer(b)
	decoded, err := message.DateFilteredReleaseFromBytes(buffer)
	if err != nil {
		t.Fatalf("DateFilteredReleaseFromBytes failed: %v", err)
	}

	if release != decoded {
		t.Fatalf("Expected %+v, got %+v", release, decoded)
	}
}
