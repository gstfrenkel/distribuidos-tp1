package review

import (
	"fmt"
	"tp1/pkg/ioutils"
	"tp1/pkg/message/positive_review"
	"tp1/pkg/message/scored_review"
	"tp1/pkg/message/utils"

	msg "tp1/pkg/message"
)

const positiveReviewScore = 1

type messages []message

type message struct {
	gameId   int64
	gameName string
	text     string
	score    int8
}

func New(gameId int64, text string, score int8) msg.Message {
	return &messages{{gameId: gameId, text: text, score: score}}
}

func FromClientBytes(c []byte) []byte {
	gameId := c[:ioutils.I64Size]
	c = c[ioutils.I64Size:]
	gameNameLen := ioutils.ReadU8FromSlice(c)
	gameName := c[:gameNameLen]
	c = c[gameNameLen:]
	textLen := ioutils.ReadU64FromSlice(c)
	text := c[:textLen]
	c = c[textLen:]
	score := uint8(ioutils.ReadI64FromSlice(c))
	_ = ioutils.ReadI64FromSlice(c) //ReviewVotes

	return append(append(append(append(append(gameId, gameNameLen), gameName...), byte(textLen)), text...), score)
}

func (ms messages) ToBytes() ([]byte, error) {
	return nil, nil
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

		key := utils.Key{GameId: m.gameId, GameName: m.gameName}

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
		key := utils.Key{GameId: m.gameId, GameName: m.gameName}

		if _, exists := reviewMsg[key]; exists {
			reviewMsg[key] = append(reviewMsg[key], m.text)
		} else {
			reviewMsg[key] = []string{m.text}
		}
	}

	return positive_review.New(reviewMsg)
}

func (m message) isPositive() bool {
	return m.score == positiveReviewScore
}
