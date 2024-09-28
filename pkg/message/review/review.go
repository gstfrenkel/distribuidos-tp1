package review

import (
	"bytes"
	"encoding/gob"
	"fmt"

	msg "tp1/pkg/message"
	"tp1/pkg/message/scored_review"
	"tp1/pkg/message/text_review"
	"tp1/pkg/message/utils"
)

const positiveReviewScore = 1

type messages []message

type message struct {
	GameId   int64
	GameName string
	Text     string
	Score    int8
}

func FromBytes(b []byte) (msg.Message, error) {
	var m messages
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return m, dec.Decode(&m)
}

func (ms messages) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(ms); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ms messages) ToMessage(id msg.ID) (msg.Message, error) {
	switch id {
	case msg.PositiveReviewID:
		return ms.toReviewMessage(false), nil
	case msg.NegativeReviewID:
		return ms.toReviewMessage(true), nil
	case msg.PositiveReviewWithTextID:
		return ms.toPositiveReviewWithTextMessage(), nil
	default:
		return nil, fmt.Errorf(msg.ErrFailedToConvert.Error(), id)
	}
}

func (ms messages) toReviewMessage(negative bool) msg.Message {
	reviewMsg := map[utils.Key]int64{}

	for _, m := range ms {
		if m.isPositive() == negative {
			continue
		}

		key := utils.Key{GameId: m.GameId, GameName: m.GameName}

		if count, exists := reviewMsg[key]; exists {
			reviewMsg[key] = count + 1
		} else {
			reviewMsg[key] = 1
		}
	}

	return scored_review.New(reviewMsg)
}

func (ms messages) toPositiveReviewWithTextMessage() msg.Message {
	reviewMsg := map[utils.Key][]string{}

	for _, m := range ms {
		key := utils.Key{GameId: m.GameId, GameName: m.GameName}

		if _, exists := reviewMsg[key]; exists {
			reviewMsg[key] = append(reviewMsg[key], m.Text)
		} else {
			reviewMsg[key] = []string{m.Text}
		}
	}

	return text_review.New(reviewMsg)
}

func (m message) isPositive() bool {
	return m.Score == positiveReviewScore
}
