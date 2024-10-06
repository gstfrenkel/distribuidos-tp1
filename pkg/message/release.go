package message

import "time"

type Releases []Release

type Release struct {
	GameId      int64
	GameName    string
	ReleaseDate string
	AvgPlaytime int64
}

func (r Releases) ToBytes() ([]byte, error) {
	return toBytes(r)
}

func ReleasesFromBytes(b []byte) (Releases, error) {
	var r Releases
	return r, fromBytes(b, &r)
}

func (r Releases) ToPlaytimeMessage(startYear int, endYear int) DateFilteredReleases {
	var result DateFilteredReleases

	layout := "Jan 2, 2006"
	for _, release := range r {
		parsedDate, err := time.Parse(layout, release.ReleaseDate)
		if err != nil {
			continue
		}

		year := parsedDate.Year()
		if year >= startYear && year <= endYear {
			filteredRelease := DateFilteredRelease{
				GameId:      release.GameId,
				GameName:    release.GameName,
				AvgPlaytime: release.AvgPlaytime,
			}
			result = append(result, filteredRelease)
		}
	}

	return result
}
