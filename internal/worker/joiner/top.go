package joiner

import (
	"fmt"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type top struct {
	w            *worker.Worker
	eofsByClient map[string]recvEofs
	gameInfoById map[string]map[int64]gameInfo
	output       amqp.Destination
}

func NewTop() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &top{
		w:            w,
		eofsByClient: map[string]recvEofs{},
		gameInfoById: map[string]map[int64]gameInfo{},
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

func (t *top) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)

	if messageId == message.EofMsg {
		processEof(
			clientId,
			delivery.Headers[amqp.OriginIdHeader].(uint8),
			t.eofsByClient,
			t.gameInfoById,
			t.processEof,
		)
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		t.processReview(clientId, msg)
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		t.processGame(clientId, msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (t *top) processEof(clientId string) {
	if err := t.w.Broker.HandleEofMessage(t.w.Id, 0, amqp.EmptyEof, headersEof, t.w.InputEof, amqp.DestinationEof(t.output)); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}

func (t *top) processReview(clientId string, msg message.ScoredReview) {
	userInfo, ok := t.gameInfoById[clientId]
	if !ok {
		t.gameInfoById[clientId] = map[int64]gameInfo{msg.GameId: {votes: msg.Votes}}
		return
	}

	info, ok := userInfo[msg.GameId]
	if !ok {
		t.gameInfoById[clientId][msg.GameId] = gameInfo{votes: msg.Votes}
		return
	}

	t.gameInfoById[clientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}

	if info.gameName == "" {
		return
	}

	b, err := message.ScoredReviews{{GameId: msg.GameId, GameName: info.gameName, Votes: info.votes + msg.Votes}}.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	headersReview[amqp.ClientIdHeader] = clientId
	if err = t.w.Broker.Publish(t.output.Exchange, t.output.Key, b, headersReview); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}

func (t *top) processGame(clientId string, msg message.GameName) {
	userInfo, ok := t.gameInfoById[clientId]
	if !ok {
		t.gameInfoById[clientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // No reviews have been received for this game.
		t.gameInfoById[clientId][msg.GameId] = gameInfo{gameName: msg.GameName}
		return
	}

	t.gameInfoById[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}

	b, err := message.ScoredReviews{{GameId: msg.GameId, Votes: info.votes, GameName: msg.GameName}}.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	headersReview[amqp.ClientIdHeader] = clientId
	if err = t.w.Broker.Publish(t.output.Exchange, t.output.Key, b, headersReview); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}
