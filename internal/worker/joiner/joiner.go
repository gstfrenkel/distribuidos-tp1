package joiner

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
)

// Every process must implement the process interface because all of them join reviews with games.
type process interface {
	processReview(headers amqp.Header, msgBytes []byte, recovery bool) []sequence.Destination
	processGame(headers amqp.Header, msgBytes []byte, recovery bool) []sequence.Destination
}

type joiner struct {
	w                *worker.Worker
	eofsByClient     map[string]recvEofs
	gameInfoByClient map[string]map[int64]gameInfo
}

type gameInfo struct {
	gameName string // If gameName is an empty string, reviews of this game have been received but the game has not yet been identified as the correct genre.
	votes    uint64
	sent     bool // Whether the game information has been forwarded to aggregator or not. Sent games should stop being processed.
}

type recvEofs struct {
	review bool
	game   bool
}

func newJoiner() (*joiner, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &joiner{
		w:                w,
		eofsByClient:     map[string]recvEofs{},
		gameInfoByClient: map[string]map[int64]gameInfo{},
	}, nil
}

func (j *joiner) processEof(header amqp.Header, sendEof func(header amqp.Header) []sequence.Destination) []sequence.Destination {
	var sequenceIds []sequence.Destination

	recv, ok := j.eofsByClient[header.ClientId]
	if !ok {
		j.eofsByClient[header.ClientId] = recvEofs{}
	}

	j.eofsByClient[header.ClientId] = recvEofs{
		review: header.OriginId == amqp.ReviewOriginId || recv.review,
		game:   header.OriginId == amqp.GameOriginId || recv.game,
	}

	if j.eofsByClient[header.ClientId].review && j.eofsByClient[header.ClientId].game {
		if sendEof != nil {
			sequenceIds = sendEof(header)
		}
		delete(j.gameInfoByClient, header.ClientId)
		delete(j.eofsByClient, header.ClientId)
	}

	return sequenceIds
}

func (j *joiner) recover(instance process) {
	ch := make(chan recovery.Message, worker.ChanSize)
	go j.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			j.processEof(recoveredMsg.Header(), nil)
		case message.ScoredReviewID:
			instance.processReview(recoveredMsg.Header(), recoveredMsg.Message(), true)
		case message.GameNameID:
			instance.processGame(recoveredMsg.Header(), recoveredMsg.Message(), true)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
