package message

import (
	"tp1/pkg/message/utils"
)

const (
	positiveReviewScore = 1
	negativeReviewScore = -1
)

type Review []review

type review struct {
	GameId   int64
	GameName string
	Text     string
	Score    int8
}

func ReviewFromBytes(b []byte) (Review, error) {
	var m Review
	return m, fromBytes(b, &m)
}

func ReviewFromClientReview(clientReview []DataCSVReviews) ([]byte, error) {
	rs := make(Review, 0, len(clientReview))
	for _, r := range clientReview {
		rs = append(rs, review{
			GameId:   r.AppID,
			GameName: r.AppName,
			Text:     r.ReviewText,
			Score:    int8(r.ReviewScore),
		})
	}

	return rs.ToBytes()
}

func (m Review) ToBytes() ([]byte, error) {
	return toBytes(m)
}

func (m Review) ToPositiveReviewMessage() ScoredReviews {
	return m.toScoredReviewMessage(positiveReviewScore)
}

func (m Review) ToNegativeReviewMessage() ScoredReviews {
	return m.toScoredReviewMessage(negativeReviewScore)
}

func (m Review) toScoredReviewMessage(targetScore int8) ScoredReviews {
	scoredReviewMsg := ScoredReviews{}
	gameVotesMap := map[int64]int64{}
	gameNamesMap := map[int64]string{}

	for _, reviewMsg := range m {
		if reviewMsg.Score != targetScore {
			continue
		}

		if count, exists := gameVotesMap[reviewMsg.GameId]; exists {
			gameVotesMap[reviewMsg.GameId] = count + 1
		} else {
			gameVotesMap[reviewMsg.GameId] = 1
			gameNamesMap[reviewMsg.GameId] = reviewMsg.GameName
		}
	}

	for k, v := range gameVotesMap {
		scoredReviewMsg = append(scoredReviewMsg, ScoredReview{GameId: k, GameName: gameNamesMap[k], Votes: v})
	}

	return scoredReviewMsg
}

func (m Review) ToPositiveReviewWithTextMessage() TextReview {
	textReviewMessage := TextReview{}

	for _, reviewMsg := range m {
		key := utils.Key{GameId: reviewMsg.GameId, GameName: reviewMsg.GameName}

		if _, exists := textReviewMessage[key]; exists {
			textReviewMessage[key] = append(textReviewMessage[key], reviewMsg.Text)
		} else {
			textReviewMessage[key] = []string{reviewMsg.Text}
		}
	}

	return textReviewMessage
}
