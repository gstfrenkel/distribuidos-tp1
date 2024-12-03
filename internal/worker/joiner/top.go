package joiner

import (
	"fmt"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

type top struct {
	joiner *joiner
	output amqp.Destination
}

func NewTop() (worker.Filter, error) {
	j, err := newJoiner()
	if err != nil {
		return nil, err
	}

	return &top{
		joiner: j,
	}, nil
}

func (t *top) Init() error {
	if err := t.joiner.w.Init(); err != nil {
		return err
	}

	t.recover()

	return nil
}

func (t *top) Start() {
	t.output = t.joiner.w.Outputs[0]
	t.output.Key = fmt.Sprintf(t.joiner.w.Outputs[0].Key, t.joiner.w.Id)

	t.joiner.w.Start(t)
}

func (t *top) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var msg []byte

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = t.joiner.processEof(headers, t.processEof)
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
	sequenceIds, err := t.joiner.w.HandleEofMessage(amqp.EmptyEof, headers, amqp.DestinationEof(t.output))
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

	userInfo, ok := t.joiner.gameInfoByClient[headers.ClientId]
	if !ok {
		t.joiner.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {votes: msg.Votes}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok {
		t.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{votes: msg.Votes}
		return sequenceIds
	}

	t.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}

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

	userInfo, ok := t.joiner.gameInfoByClient[headers.ClientId]
	if !ok {
		t.joiner.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // No reviews have been received for this game.
		t.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName}
		return sequenceIds
	}

	t.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}

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
	sequenceId := t.joiner.w.NextSequenceId(key)

	headers = headers.WithMessageId(message.ScoredReviewID).WithSequenceId(sequence.SrcNew(t.joiner.w.Uuid, sequenceId))

	if err := t.joiner.w.Broker.Publish(t.output.Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		return nil
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func (t *top) recover() {
	t.joiner.recover(t)
}
