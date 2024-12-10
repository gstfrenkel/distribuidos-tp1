package message

import (
	"bytes"
	"fmt"
	"strings"
	"tp1/pkg/utils/encoding"
)

type DateFilteredReleases []DateFilteredRelease

type DateFilteredRelease struct {
	GameId      int64
	GameName    string
	AvgPlaytime int64
}

func (r DateFilteredReleases) ToBytes() ([]byte, error) {
	size := uint32(len(r))
	sizeBuff := bytes.Buffer{}
	if err := encoding.EncodeNumber(&sizeBuff, size); err != nil {
		return nil, err
	}

	b := make([]byte, 0, size)
	b = append(b, sizeBuff.Bytes()...)

	for _, rs := range r {
		releaseBytes, err := rs.ToBytes()
		if err != nil {
			return nil, err
		}

		b = append(b, releaseBytes...)
	}

	return b, nil
}

func (r DateFilteredRelease) ToBytes() ([]byte, error) {
	buff := bytes.Buffer{}

	if err := encoding.EncodeNumber(&buff, r.GameId); err != nil {
		return nil, err
	}

	if err := encoding.EncodeString(&buff, r.GameName); err != nil {
		return nil, err
	}

	if err := encoding.EncodeNumber(&buff, r.AvgPlaytime); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func DateFilteredReleaseFromBytes(buff *bytes.Buffer) (DateFilteredRelease, error) {
	gameId, err := encoding.DecodeInt64(buff)
	if err != nil {
		return DateFilteredRelease{}, err
	}

	gameName, err := encoding.DecodeString(buff)
	if err != nil {
		return DateFilteredRelease{}, err
	}

	avgPlaytime, err := encoding.DecodeInt64(buff)
	if err != nil {
		return DateFilteredRelease{}, err
	}

	return DateFilteredRelease{gameId, gameName, avgPlaytime}, nil
}

func DateFilteredReleasesFromBytes(b []byte) (DateFilteredReleases, error) {
	var r DateFilteredReleases
	buff := bytes.NewBuffer(b)
	size, err := encoding.DecodeUint32(buff)
	if err != nil {
		return nil, err
	}

	for i := uint32(0); i < size; i++ {
		release, err := DateFilteredReleaseFromBytes(buff)
		if err != nil {
			return nil, err
		}

		r = append(r, release)
	}

	return r, nil
}

func (r DateFilteredReleases) ToResultString() string {
	header := "Q2:\n"
	var gamesInfo []string
	for _, release := range r {
		gamesInfo = append(gamesInfo, fmt.Sprintf("Juego: [%s], AvgPlaytime: [%d]", release.GameName, release.AvgPlaytime))
	}
	return header + strings.Join(gamesInfo, "\n")
}
