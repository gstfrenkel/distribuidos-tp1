package review

import (
	"bytes"
	"encoding/gob"
	"tp1/pkg/message/scored_review"
	"tp1/pkg/message/text_review"
	"tp1/pkg/message/utils"
)

const (
	positiveReviewScore = 1
	negativeReviewScore = -1
)

type Message []struct {
	GameId   int64
	GameName string
	Text     string
	Score    int8
}

func FromBytes(b []byte) (Message, error) {
	var m Message
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return m, dec.Decode(&m)
}

func (m Message) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m Message) ToPositiveReviewMessage() scored_review.Messages {
	return m.toScoredReviewMessage(positiveReviewScore)
}

func (m Message) ToNegativeReviewMessage() scored_review.Messages {
	return m.toScoredReviewMessage(negativeReviewScore)
}

func (m Message) toScoredReviewMessage(targetScore int8) scored_review.Messages {
	scoredReviewMsg := scored_review.Messages{}
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
		scoredReviewMsg = append(scoredReviewMsg, scored_review.Message{GameId: k, GameName: gameNamesMap[k], Votes: v})
	}

	return scoredReviewMsg
}

func (m Message) ToPositiveReviewWithTextMessage() text_review.Message {
	textReviewMessage := text_review.Message{}

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
