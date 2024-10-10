package message

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
