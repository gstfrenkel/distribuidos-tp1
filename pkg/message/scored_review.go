package message

type ScoredReviews []ScoredReview

type ScoredReview struct {
	GameId   int64
	Votes    uint64
	GameName string
}

func ScoredReviewFromBytes(b []byte) (ScoredReview, error) {
	var m ScoredReview
	return m, fromBytes(b, &m)
}

func ScoredReviewsFromBytes(b []byte) (ScoredReviews, error) {
	var m ScoredReviews
	return m, fromBytes(b, &m)
}

func (m ScoredReview) ToBytes() ([]byte, error) {
	return toBytes(m)
}

func (m ScoredReviews) ToBytes() ([]byte, error) {
	return toBytes(m)
}

func ScoredRevFromAnyToBytes(s []any) ([]byte, error) {
	var m ScoredReviews
	for _, review := range s {
		m = append(m, review.(ScoredReview))
	}
	return toBytes(m)
}

func (m ScoredReviews) ToAny() []any {
	var dataAny []any
	for _, review := range m {
		dataAny = append(dataAny, review)
	}
	return dataAny
}
