package message

import (
	"fmt"
	"strings"
)

type DateFilteredReleases []DateFilteredRelease
type DateFilteredRelease struct {
	GameId      int64
	GameName    string
	AvgPlaytime int64
}

func (r DateFilteredReleases) ToBytes() ([]byte, error) {
	return toBytes(r)
}

func DateFilteredReleasesFromBytes(b []byte) (DateFilteredReleases, error) {
	var r DateFilteredReleases
	return r, fromBytes(b, &r)
}

func (releases DateFilteredReleases) ToResultString() string {
	header := "Q2:\n"
	var gamesInfo []string
	for _, release := range releases {
		gamesInfo = append(gamesInfo, fmt.Sprintf("Juego: [%s], AvgPlaytime: [%d]", release.GameName, release.AvgPlaytime))
	}
	return header + strings.Join(gamesInfo, "\n")
}
