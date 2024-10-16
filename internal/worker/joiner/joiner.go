package joiner

import (
	"tp1/pkg/amqp"
	"tp1/pkg/message"
)

var (
	headersEof    = map[string]any{amqp.OriginIdHeader: amqp.ReviewOriginId, amqp.MessageIdHeader: uint8(message.EofMsg)}
	headersGame   = map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID)}
	headersReview = map[string]any{amqp.MessageIdHeader: uint8(message.ScoredReviewID)}
)

type counterGameInfo struct {
	gameName string // If gameName is an empty string, reviews of this game have been received but the game has not yet been identified as the correct genre.
	votes    uint64
	sent     bool // Whether the game information has been forwarded to aggregator or not. Sent games should stop being processed.
}

type gameInfo struct {
	gameName string // If gameName is an empty string, reviews of this game have been received but the game has not yet been identified as the correct genre.
	votes    uint64
}
