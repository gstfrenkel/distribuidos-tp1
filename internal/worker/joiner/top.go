package joiner

import (
	"fmt"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
)

type top struct {
	w                *worker.Worker
	eofsByClient     map[string]recvEofs
	gameInfoByClient map[string]map[int64]gameInfo
	output           amqp.Destination
}

func NewTop() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &top{
		w:                w,
		eofsByClient:     map[string]recvEofs{},
		gameInfoByClient: map[string]map[int64]gameInfo{},
	}, nil
}

func (t *top) Init() error {
	return t.w.Init()
}

func (t *top) Start() {
	t.output = t.w.Outputs[0]
	t.output.Key = fmt.Sprintf(t.w.Outputs[0].Key, t.w.Id)

	t.w.Start(t)
}

func (t *top) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var msg []byte

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = processEof(
			headers,
			t.eofsByClient,
			t.gameInfoByClient,
			t.processEof,
		)
	case message.ScoredReviewID:
		sequenceIds = t.processReview(headers, delivery.Body, false)
	case message.GameNameID:
		sequenceIds = t.processGame(headers, delivery.Body, false)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, msg
}

func (t *top) processEof(headers amqp.Header) []sequence.Destination {
	sequenceIds, err := t.w.HandleEofMessage(amqp.EmptyEof, headers, amqp.DestinationEof(t.output))
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequenceIds
}

func (t *top) processReview(headers amqp.Header, msgBytes []byte, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	msg, err := message.ScoredReviewFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return sequenceIds
	}

	userInfo, ok := t.gameInfoByClient[headers.ClientId]
	if !ok {
		t.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {votes: msg.Votes}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok {
		t.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{votes: msg.Votes}
		return sequenceIds
	}

	t.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}

	if info.gameName == "" {
		return sequenceIds
	}

	if !recovery {
		b, err := message.ScoredReviews{{GameId: msg.GameId, GameName: info.gameName, Votes: info.votes + msg.Votes}}.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return sequenceIds
		}
		destSeqId := t.publish(headers, b)
		if destSeqId == nil {
			return sequenceIds
		}
		sequenceIds = append(sequenceIds, destSeqId...)
	}

	return sequenceIds
}

func (t *top) processGame(headers amqp.Header, msgBytes []byte, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	msg, err := message.GameNameFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return sequenceIds
	}

	userInfo, ok := t.gameInfoByClient[headers.ClientId]
	if !ok {
		t.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // No reviews have been received for this game.
		t.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName}
		return sequenceIds
	}

	t.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}

	if !recovery {
		b, err := message.ScoredReviews{{GameId: msg.GameId, Votes: info.votes, GameName: msg.GameName}}.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return sequenceIds
		}
		destSeqId := t.publish(headers, b)
		if destSeqId == nil {
			return sequenceIds
		}
		sequenceIds = append(sequenceIds, destSeqId...)
	}

	return sequenceIds
}

func (t *top) publish(headers amqp.Header, b []byte) []sequence.Destination {
	key := t.output.Key
	sequenceId := t.w.NextSequenceId(key)

	headers = headers.WithMessageId(message.ScoredReviewID).WithSequenceId(sequence.SrcNew(t.w.Id, sequenceId))

	if err := t.w.Broker.Publish(t.output.Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		return nil
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func (t *top) recover() {
	ch := make(chan recovery.Message, worker.ChanSize)
	go t.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			processEof(
				recoveredMsg.Header(),
				t.eofsByClient,
				t.gameInfoByClient,
				nil,
			)
		case message.ScoredReviewID:
			t.processReview(recoveredMsg.Header(), recoveredMsg.Message(), true)
		case message.GameNameID:
			t.processGame(recoveredMsg.Header(), recoveredMsg.Message(), true)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
