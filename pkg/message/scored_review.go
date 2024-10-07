package message

type ScoredReviews []ScoredReview
type GameId int64

type ScoredReview struct {
	GameId   GameId
	Votes    uint64
	GameName string
}

func ScoredReviewFromBytes(b []byte) (ScoredReview, error) {
	var m ScoredReview
	return m, fromBytes(b, &m)
}

func (m ScoredReview) ToBytes() ([]byte, error) {
	return toBytes(m)
}

func (m ScoredReviews) ToBytes() ([]byte, error) {
	return toBytes(m)
}
