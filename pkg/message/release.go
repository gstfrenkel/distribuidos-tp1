package message

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
