package top_n_playtime

import (
	"testing"
	"tp1/pkg/message"
)

func TestTop10ReleasesWith13Games(t *testing.T) {
	games := message.DateFilteredReleases{
		{GameId: 1, GameName: "Game 1", AvgPlaytime: 100},
		{GameId: 2, GameName: "Game 2", AvgPlaytime: 200},
		{GameId: 3, GameName: "Game 3", AvgPlaytime: 50},
		{GameId: 4, GameName: "Game 4", AvgPlaytime: 300},
		{GameId: 5, GameName: "Game 5", AvgPlaytime: 400},
		{GameId: 6, GameName: "Game 6", AvgPlaytime: 500},
		{GameId: 7, GameName: "Game 7", AvgPlaytime: 250},
		{GameId: 8, GameName: "Game 8", AvgPlaytime: 150},
		{GameId: 9, GameName: "Game 9", AvgPlaytime: 700},
		{GameId: 10, GameName: "Game 10", AvgPlaytime: 600},
		{GameId: 11, GameName: "Game 11", AvgPlaytime: 800},
		{GameId: 12, GameName: "Game 12", AvgPlaytime: 350},
		{GameId: 13, GameName: "Game 13", AvgPlaytime: 450},
	}

	h := &MinHeapPlaytime{}
	h.UpdateReleases(games, 10)

	topReleases := h.GetTopReleases()

	expected := message.DateFilteredReleases{
		{GameId: 11, GameName: "Game 11", AvgPlaytime: 800},
		{GameId: 9, GameName: "Game 9", AvgPlaytime: 700},
		{GameId: 10, GameName: "Game 10", AvgPlaytime: 600},
		{GameId: 6, GameName: "Game 6", AvgPlaytime: 500},
		{GameId: 13, GameName: "Game 13", AvgPlaytime: 450},
		{GameId: 5, GameName: "Game 5", AvgPlaytime: 400},
		{GameId: 12, GameName: "Game 12", AvgPlaytime: 350},
		{GameId: 4, GameName: "Game 4", AvgPlaytime: 300},
		{GameId: 7, GameName: "Game 7", AvgPlaytime: 250},
		{GameId: 2, GameName: "Game 2", AvgPlaytime: 200},
	}

	if len(topReleases) != 10 {
		t.Errorf("Expected 10 top releases, got %d", len(topReleases))
	}

	for i := range expected {
		if topReleases[i] != expected[i] {
			t.Errorf("At index %d, expected %v, got %v", i, expected[i], topReleases[i])
		}
	}
}

func TestTop10ReleasesWith8Games(t *testing.T) {
	games := message.DateFilteredReleases{
		{GameId: 1, GameName: "Game 1", AvgPlaytime: 100},
		{GameId: 2, GameName: "Game 2", AvgPlaytime: 200},
		{GameId: 3, GameName: "Game 3", AvgPlaytime: 50},
		{GameId: 4, GameName: "Game 4", AvgPlaytime: 300},
		{GameId: 5, GameName: "Game 5", AvgPlaytime: 400},
		{GameId: 6, GameName: "Game 6", AvgPlaytime: 500},
		{GameId: 7, GameName: "Game 7", AvgPlaytime: 250},
		{GameId: 8, GameName: "Game 8", AvgPlaytime: 150},
	}

	h := &MinHeapPlaytime{}
	h.UpdateReleases(games, 10)

	topReleases := h.GetTopReleases()

	expected := message.DateFilteredReleases{
		{GameId: 6, GameName: "Game 6", AvgPlaytime: 500},
		{GameId: 5, GameName: "Game 5", AvgPlaytime: 400},
		{GameId: 4, GameName: "Game 4", AvgPlaytime: 300},
		{GameId: 7, GameName: "Game 7", AvgPlaytime: 250},
		{GameId: 2, GameName: "Game 2", AvgPlaytime: 200},
		{GameId: 8, GameName: "Game 8", AvgPlaytime: 150},
		{GameId: 1, GameName: "Game 1", AvgPlaytime: 100},
		{GameId: 3, GameName: "Game 3", AvgPlaytime: 50},
	}

	if len(topReleases) != 8 {
		t.Errorf("Expected 8 top releases, got %d", len(topReleases))
	}

	for i := range expected {
		if topReleases[i] != expected[i] {
			t.Errorf("At index %d, expected %v, got %v", i, expected[i], topReleases[i])
		}
	}
}

