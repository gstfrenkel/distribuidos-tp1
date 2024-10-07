package joiner

import (
	"fmt"

	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type topGameInfo struct {
	gameName string // If gameName is an empty string, reviews of this game have been received but the game has not yet been identified as the correct genre.
	votes    int64
}

type top struct {
	w             *worker.Worker
	recvReviewEof bool
	recvGameEof   bool
	gameInfoById  map[int64]topGameInfo
	outputKey     string
}

func NewTop() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &top{w: w, gameInfoById: map[int64]topGameInfo{}}, nil
}

func (t *top) Init() error {
	return t.w.Init()
}

func (t *top) Start() {
	t.outputKey = fmt.Sprintf(t.w.Outputs[0].Key, t.w.Id)

	t.w.Start(t)
}

func (t *top) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		t.processEof(delivery.Headers[amqp.OriginIdHeader].(uint8))
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		t.processReview(msg)
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		t.processGame(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (t *top) processEof(origin uint8) {
	if origin == amqp.GameOriginId {
		t.recvReviewEof = true
	} else if origin == amqp.ReviewOriginId {
		t.recvReviewEof = true
	} else {
		logs.Logger.Errorf(fmt.Sprintf("Unknown message origin ID received: %d", origin))
	}

	if t.recvReviewEof && t.recvGameEof {
		headers := map[string]any{amqp.OriginIdHeader: amqp.GameOriginId}
		destination := t.w.Outputs[0]
		destination.Key = t.outputKey

		if err := t.w.Broker.HandleEofMessage(t.w.Id, t.w.Peers, message.Eof{}, headers, t.w.InputEof, amqp.DestinationEof(t.w.Outputs[0])); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}

		t.recvReviewEof = false
		t.recvGameEof = false
		t.gameInfoById = map[int64]topGameInfo{}
	}
}

func (t *top) processReview(msg message.ScoredReview) {
	info, ok := t.gameInfoById[msg.GameId]
	if !ok {
		t.gameInfoById[msg.GameId] = topGameInfo{votes: msg.Votes}
		return
	}

	t.gameInfoById[msg.GameId] = topGameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}

	if info.gameName == "" {
		return
	}

	b, err := message.ScoredReview{GameId: msg.GameId, GameName: info.gameName, Votes: info.votes + msg.Votes}.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else if err = t.w.Broker.Publish(t.w.Outputs[0].Exchange, t.outputKey, b, map[string]any{amqp.MessageIdHeader: message.GameNameID}); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}

func (t *top) processGame(msg message.GameName) {
	info, ok := t.gameInfoById[msg.GameId]
	if !ok { // No reviews have been received for this game.
		t.gameInfoById[msg.GameId] = topGameInfo{gameName: msg.GameName}
		return
	}

	t.gameInfoById[msg.GameId] = topGameInfo{gameName: msg.GameName, votes: info.votes}

	b, err := message.ScoredReview{GameId: msg.GameId, Votes: info.votes, GameName: msg.GameName}.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else if err = t.w.Broker.Publish(t.w.Outputs[0].Exchange, t.outputKey, b, map[string]any{amqp.MessageIdHeader: message.GameNameID}); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}
