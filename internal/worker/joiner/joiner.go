package joiner

import (
	"tp1/pkg/amqp"
	"tp1/pkg/sequence"
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

func processEof(header amqp.Header, eofsByClient map[string]recvEofs, gameInfoByClient map[string]map[int64]gameInfo, f func(header amqp.Header) []sequence.Destination) []sequence.Destination {
	var sequenceIds []sequence.Destination

	recv, ok := eofsByClient[header.ClientId]
	if !ok {
		eofsByClient[header.ClientId] = recvEofs{}
	}

	eofsByClient[header.ClientId] = recvEofs{
		review: header.OriginId == amqp.ReviewOriginId || recv.review,
		game:   header.OriginId == amqp.GameOriginId || recv.game,
	}

	if eofsByClient[header.ClientId].review && eofsByClient[header.ClientId].game {
		if f != nil {
			sequenceIds = f(header)
		}
		delete(gameInfoByClient, header.ClientId)
		delete(eofsByClient, header.ClientId)
	}

	return sequenceIds
}
