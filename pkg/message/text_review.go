package message

type TextReviews map[int64][]string

func TextReviewFromBytes(b []byte) (TextReviews, error) {
	var m TextReviews
	return m, fromBytes(b, &m)
}

func (m TextReviews) ToBytes() ([]byte, error) {
	return toBytes(m)
}
