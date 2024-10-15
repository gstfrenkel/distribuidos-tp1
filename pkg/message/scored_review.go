package message

import (
	"fmt"
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

func ScoredRevFromAnyToBytes(s []any) ([]byte, error) {
	var m ScoredReviews
	for _, review := range s {
		m = append(m, review.(ScoredReview))
	}
	return toBytes(m)
}

func (reviews ScoredReviews) ToQ3ResultString() string {
	header := fmt.Sprintf("Q3: Juegos top 5 del género Indie con más reseñas positivas\n")
	var reviewsInfo []string
	for _, review := range reviews {
		reviewsInfo = append(reviewsInfo, fmt.Sprintf("Juego: [%s], Reseñas positivas: [%d] \n", review.GameName, review.Votes))
	}
	return header + strings.Join(reviewsInfo, "")
}

func ToQ5ResultString(reviewsInfo string) string {
	header := fmt.Sprintf("Q5: juegos del género Action dentro del percentil 90 en cantidad de reseñas negativas\n")
	return header + reviewsInfo
}

func (reviews ScoredReviews) ToStringAux() string {
	var reviewsInfo []string
	for _, review := range reviews {
		reviewsInfo = append(reviewsInfo, fmt.Sprintf("Juego: [%s], Reseñas negativas: [%d] \n", review.GameName, review.Votes))
	}
	return strings.Join(reviewsInfo, "")
}

func (m ScoredReviews) ToAny() []any {
	var dataAny []any
	for _, review := range m {
		dataAny = append(dataAny, review)
	}
	return dataAny
}
