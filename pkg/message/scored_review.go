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

func (m ScoredReviews) ToGameNameBytes() ([]byte, error) {
	var gameNames GameNames
	for _, scoredReview := range m {
		gameNames = append(gameNames, GameName{GameId: scoredReview.GameId, GameName: scoredReview.GameName})
	}

	if gameNames != nil {
		return gameNames.ToBytes()
	}
	return nil, nil
}

func (reviews ScoredReviews) ToQ3ResultString() string {
	header := fmt.Sprintf("Q3: Juegos top 5 del género Indie con más reseñas positivas\n")
	var reviewsInfo []string
	for _, review := range reviews {
		reviewsInfo = append(reviewsInfo, fmt.Sprintf("Juego: [%s], Reseñas positivas: [%d] \n", review.GameName, review.Votes))
	}
	return header + strings.Join(reviewsInfo, "")
}

func (reviews ScoredReviews) ToQ5ResultString() string {
	header := fmt.Sprintf("Q5:juegos del género Action dentro del percentil 90 en cantidad de reseñas negativas\n")
	var reviewsInfo []string
	for _, review := range reviews {
		reviewsInfo = append(reviewsInfo, fmt.Sprintf("Juego: [%s], Reseñas negativas: [%d] \n", review.GameName, review.Votes))
	}
	return header + strings.Join(reviewsInfo, "")
}
