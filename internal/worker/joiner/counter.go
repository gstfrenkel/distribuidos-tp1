package joiner

import (
	"fmt"

	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type counterGameInfo struct {
	gameName string // If gameName is an empty string, reviews of this game have been received but the game has not yet been identified as the correct genre.
	votes    uint64
	sent     bool // Whether the game information has been forwarded to aggregator or not. Sent games should stop being processed.
}

type counter struct {
	w             *worker.Worker
	recvReviewEof bool
	recvGameEof   bool
	target        uint64
	gameInfoById  map[int64]counterGameInfo
}

func NewCounter() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &counter{w: w, gameInfoById: map[int64]counterGameInfo{}}, nil
}

func (c *counter) Init() error {
	return c.w.Init()
}

func (c *counter) Start() {
	c.target = uint64(c.w.Query.(float64))

	c.w.Start(c)
}

func (c *counter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		c.processEof(delivery.Headers[amqp.OriginIdHeader].(uint8))
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		c.processReview(msg)
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		c.processGame(msg, delivery.Body)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (c *counter) processEof(origin uint8) {
	if origin == amqp.GameOriginId {
		c.recvGameEof = true
	} else if origin == amqp.ReviewOriginId {
		c.recvReviewEof = true
	} else {
		logs.Logger.Errorf(fmt.Sprintf("Unknown message origin ID received: %d", origin))
	}

	if c.recvReviewEof && c.recvGameEof {
		headers := map[string]any{amqp.OriginIdHeader: amqp.GameOriginId}

		if err := c.w.Broker.HandleEofMessage(c.w.Id, 0, message.Eof{}, headers, c.w.InputEof, c.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}

		c.recvReviewEof = false
		c.recvGameEof = false
		c.gameInfoById = map[int64]counterGameInfo{}
	}
}

func (c *counter) processReview(msg message.ScoredReview) {
	info, ok := c.gameInfoById[msg.GameId]
	if !ok {
		c.gameInfoById[msg.GameId] = counterGameInfo{votes: msg.Votes}
		return
	} else if info.sent {
		return
	}

	if info.gameName != "" && info.votes+msg.Votes >= c.target {
		b, err := message.GameName{GameId: msg.GameId, GameName: info.gameName}.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
			return
		}

		if err = c.w.Broker.Publish(c.w.Outputs[0].Exchange, c.w.Outputs[0].Key, b, map[string]any{amqp.MessageIdHeader: message.GameNameID}); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		} else {
			c.gameInfoById[msg.GameId] = counterGameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
		}
	} else if info.votes < c.target {
		c.gameInfoById[msg.GameId] = counterGameInfo{votes: info.votes + msg.Votes}
	}
}

func (c *counter) processGame(msg message.GameName, b []byte) {
	info, ok := c.gameInfoById[msg.GameId]
	if !ok { // No reviews have been received for this game.
		c.gameInfoById[msg.GameId] = counterGameInfo{gameName: msg.GameName}
		return
	}

	if info.votes >= c.target { // Reviews have been received, and they exceed the target vote count.
		if err := c.w.Broker.Publish(c.w.Outputs[0].Exchange, c.w.Outputs[0].Key, b, map[string]any{amqp.MessageIdHeader: message.GameNameID}); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
			return
		}

		c.gameInfoById[msg.GameId] = counterGameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
	} else { // Reviews have been received, but they do not exceed the target vote count.
		c.gameInfoById[msg.GameId] = counterGameInfo{gameName: msg.GameName, votes: info.votes}
	}
}
