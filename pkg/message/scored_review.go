package message

type ScoredReviews []ScoredReview

type ScoredReview struct {
	GameId   int64
	Votes    int64
	GameName string
}

func ScoredReviewFromBytes(b []byte) (ScoredReview, error) {
	var m ScoredReview
	return m, fromBytes(b, &m)
}

func (m ScoredReview) ToBytes() ([]byte, error) {
	return toBytes(m)
}
