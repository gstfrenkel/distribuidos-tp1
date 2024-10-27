package joiner

import (
	"tp1/pkg/amqp"
	"tp1/pkg/message"
)

var (
	headersEof    = map[string]any{amqp.MessageIdHeader: uint8(message.EofMsg)}
	headersGame   = map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID)}
	headersReview = map[string]any{amqp.MessageIdHeader: uint8(message.ScoredReviewID)}
)

type gameInfo struct {
	gameName string // If gameName is an empty string, reviews of this game have been received but the game has not yet been identified as the correct genre.
	votes    uint64
	sent     bool // Whether the game information has been forwarded to aggregator or not. Sent games should stop being processed.
}

type recvEofs struct {
	review bool
	game   bool
}

func processEof(clientId string, origin uint8, eofsByClient map[string]recvEofs, gameInfoByClient map[string]map[int64]gameInfo, f func(clientId string)) {
	recv, ok := eofsByClient[clientId]
	if !ok {
		eofsByClient[clientId] = recvEofs{}
	}

	eofsByClient[clientId] = recvEofs{
		review: origin == amqp.ReviewOriginId || recv.review,
		game:   origin == amqp.GameOriginId || recv.game,
	}

	if eofsByClient[clientId].review && eofsByClient[clientId].game {
		headersEof[amqp.ClientIdHeader] = clientId
		f(clientId)
		delete(gameInfoByClient, clientId)
		delete(eofsByClient, clientId)
	}
}
