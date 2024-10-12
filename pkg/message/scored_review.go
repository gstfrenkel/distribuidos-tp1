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

func ToGameNameBytes(m any) ([]byte, error) {
	s := m.(ScoredReviews)
	var gameNames GameNames
	for _, scoredReview := range s {
		gameNames = append(gameNames, GameName{GameId: scoredReview.GameId, GameName: scoredReview.GameName})
	}

	if gameNames != nil {
		return gameNames.ToBytes()
	}
	return nil, nil
}

func (m ScoredReviews) ToAny() []any {
	var dataAny []any
	for _, review := range m {
		dataAny = append(dataAny, review)
	}
	return dataAny
}
