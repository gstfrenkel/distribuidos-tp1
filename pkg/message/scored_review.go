package message

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"tp1/pkg/utils/encoding"
)

type ScoredReviews []ScoredReview

type ScoredReview struct {
	GameId   int64
	Votes    uint64
	GameName string
}

func ScoredReviewFromBytes(b []byte) (ScoredReview, error) {
	return scoredReviewFromBuf(bytes.NewBuffer(b))
}

func ScoredReviewsFromBytes(b []byte) (ScoredReviews, error) {
	buf := bytes.NewBuffer(b)

	size, err := encoding.DecodeUint32(buf)
	if err != nil {
		return nil, err
	}

	scoredReviews := make(ScoredReviews, 0, size)
	for i := uint32(0); i < size; i++ {
		scoredReview, err := scoredReviewFromBuf(buf)
		if err != nil {
			return nil, err
		}
		scoredReviews = append(scoredReviews, scoredReview)
	}

	return scoredReviews, nil
}

func scoredReviewFromBuf(buf *bytes.Buffer) (ScoredReview, error) {
	id, err := encoding.DecodeInt64(buf)
	if err != nil {
		return ScoredReview{}, err
	}

	votes, err := encoding.DecodeUint64(buf)
	if err != nil {
		return ScoredReview{}, err
	}

	name, err := encoding.DecodeString(buf)
	if err != nil {
		return ScoredReview{}, err
	}

	return ScoredReview{
		GameId:   id,
		Votes:    votes,
		GameName: name,
	}, nil
}

func (m ScoredReview) ToBytes() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	if err := encoding.EncodeNumber(buf, m.GameId); err != nil {
		return nil, err
	}

	if err := encoding.EncodeNumber(buf, m.Votes); err != nil {
		return nil, err
	}

	if err := encoding.EncodeString(buf, m.GameName); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m ScoredReviews) ToBytes() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	if err := encoding.EncodeNumber(buf, uint32(len(m))); err != nil {
		return nil, err
	}

	b := buf.Bytes()
	for _, scoredReview := range m {
		aux, err := scoredReview.ToBytes()
		if err != nil {
			return nil, err
		}
		b = append(b, aux...)
	}

	return b, nil
}

func (m ScoredReviews) ToQ3ResultString() string {
	header := fmt.Sprintf("Q3:\n")
	var reviewsInfo []string
	for _, r := range m {
		reviewsInfo = append(reviewsInfo, fmt.Sprintf("Juego: [%s], Reseñas positivas: [%d]", r.GameName, r.Votes))
	}
	return header + strings.Join(reviewsInfo, "\n")
}

func (m ScoredReviews) Sort(ascending bool) {
	sort.Slice(m, func(i, j int) bool {
		if m[i].Votes != m[j].Votes {
			return (ascending && m[i].Votes < m[j].Votes) || (!ascending && m[i].Votes > m[j].Votes)
		}

		return m[i].GameId < m[j].GameId
	})
}

func ToQ5ResultString(reviewsInfo string) string {
	header := fmt.Sprintf("Q5:\n")
	return header + reviewsInfo
}

func (m ScoredReviews) ToStringAux() string {
	var reviewsInfo []string
	for _, r := range m {
		reviewsInfo = append(reviewsInfo, fmt.Sprintf("Juego: [%s], Reseñas negativas: [%d]", r.GameName, r.Votes))
	}
	return strings.Join(reviewsInfo, "\n") + "\n"
}
