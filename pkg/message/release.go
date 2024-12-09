package message

import (
	"bytes"
	"time"
	"tp1/pkg/utils/encoding"
)

type Releases []release

type release struct {
	GameId      int64
	GameName    string
	ReleaseDate string
	AvgPlaytime int64
}

func (r Releases) ToBytes() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	if err := encoding.EncodeNumber(buf, uint32(len(r))); err != nil {
		return nil, err
	}

	for _, rev := range r {
		if err := encoding.EncodeNumber(buf, rev.GameId); err != nil {
			return nil, err
		}

		if err := encoding.EncodeString(buf, rev.GameName); err != nil {
			return nil, err
		}

		if err := encoding.EncodeString(buf, rev.ReleaseDate); err != nil {
			return nil, err
		}

		if err := encoding.EncodeNumber(buf, rev.AvgPlaytime); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func ReleasesFromBytes(b []byte) (Releases, error) {
	buf := bytes.NewBuffer(b)

	size, err := encoding.DecodeUint32(buf)
	if err != nil {
		return nil, err
	}

	releases := make([]release, 0, size)
	for i := uint32(0); i < size; i++ {
		id, err := encoding.DecodeInt64(buf)
		if err != nil {
			return nil, err
		}

		name, err := encoding.DecodeString(buf)
		if err != nil {
			return nil, err
		}

		date, err := encoding.DecodeString(buf)
		if err != nil {
			return nil, err
		}

		playtime, err := encoding.DecodeInt64(buf)
		if err != nil {
			return nil, err
		}

		releases = append(releases, release{
			GameId:      id,
			GameName:    name,
			ReleaseDate: date,
			AvgPlaytime: playtime,
		})
	}

	return releases, nil
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
