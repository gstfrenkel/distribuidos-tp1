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

func (t *top) Process(delivery amqp.Delivery, header amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var msg []byte

	switch header.MessageId {
	case message.EofMsg:
		sequenceIds = processEof(
			header,
			t.eofsByClient,
			t.gameInfoByClient,
			t.processEof,
		)
	case message.ScoredReviewID:
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = t.processReview(header.ClientId, msg)
		}
	case message.GameNameID:
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = t.processGame(header.ClientId, msg)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), header.MessageId)
	}

	return sequenceIds, msg
}

func (t *top) processEof(_ string) []sequence.Destination {
	sequenceIds, err := t.w.HandleEofMessage(amqp.EmptyEof, headersEof, amqp.DestinationEof(t.output))
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequenceIds
}

func (t *top) processReview(clientId string, msg message.ScoredReview) []sequence.Destination {
	var sequenceIds []sequence.Destination

	userInfo, ok := t.gameInfoByClient[clientId]
	if !ok {
		t.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {votes: msg.Votes}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok {
		t.gameInfoByClient[clientId][msg.GameId] = gameInfo{votes: msg.Votes}
		return sequenceIds
	}

	t.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}

	if info.gameName == "" {
		return sequenceIds
	}

	b, err := message.ScoredReviews{{GameId: msg.GameId, GameName: info.gameName, Votes: info.votes + msg.Votes}}.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	key := t.output.Key
	sequenceId := t.w.NextSequenceId(key)
	sequenceIds = append(sequenceIds, sequence.DstNew(key, sequenceId))

	headersReview[amqp.SequenceIdHeader] = sequence.SrcNew(t.w.Id, sequenceId)
	headersReview[amqp.ClientIdHeader] = clientId
	if err = t.w.Broker.Publish(t.output.Exchange, key, b, headersReview); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequenceIds
}

func (t *top) processGame(clientId string, msg message.GameName) []sequence.Destination {
	var sequenceIds []sequence.Destination

	userInfo, ok := t.gameInfoByClient[clientId]
	if !ok {
		t.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // No reviews have been received for this game.
		t.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName}
		return sequenceIds
	}

	t.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}

	b, err := message.ScoredReviews{{GameId: msg.GameId, Votes: info.votes, GameName: msg.GameName}}.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	key := t.output.Key
	sequenceId := t.w.NextSequenceId(key)
	sequenceIds = append(sequenceIds, sequence.DstNew(key, sequenceId))

	headersReview[amqp.SequenceIdHeader] = sequence.SrcNew(t.w.Id, sequenceId)
	headersReview[amqp.ClientIdHeader] = clientId
	if err = t.w.Broker.Publish(t.output.Exchange, key, b, headersReview); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequenceIds
}
