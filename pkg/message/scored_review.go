package message

import (
	"fmt"
	"sort"
	"strings"
)

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
	return strings.Join(reviewsInfo, "\n")
}
