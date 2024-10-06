package joiner

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type gameInfo struct {
	gameName string // If gameName is an empty string, reviews of this game have been received but the game has not yet been identified as the correct genre.
	votes    int64
	sent     bool // Whether the game information has been forwarded to aggregator or not. Sent games should stop being processed.
}

type counter struct {
	w             *worker.Worker
	recvReviewEof bool
	recvGameEof   bool
	target        int64
	gameInfoById  map[int64]gameInfo
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &counter{w: w}, nil
}

func (c *counter) Init() error {
	return c.w.Init()
}

func (c *counter) Start() {
	c.target = int64(c.w.Query.(float64))

	c.w.Start(c)
}

func (c *counter) Process(reviewDelivery amqp.Delivery) {
	messageId := message.ID(reviewDelivery.Headers[amqp.MessageIdHeader].(uint8))
	headers := map[string]any{amqp.OriginIdHeader: amqp.GameOriginId}

	if messageId == message.EofMsg {

	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(reviewDelivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		c.processReview(msg)
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(reviewDelivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		c.processGame(msg, b)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (c *counter) processReview(msg message.ScoredReview) {
	info, ok := c.gameInfoById[msg.GameId]
	if !ok {
		c.gameInfoById[msg.GameId] = gameInfo{votes: msg.Votes}
		return
	}

	if info.sent {
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
			c.gameInfoById[msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
		}
	} else if info.gameName == "" && info.votes < c.target {

	}
}

func (c *counter) processGame(msg message.GameName, b []byte) {
	info, ok := c.gameInfoById[msg.GameId]
	if !ok { // No reviews have been received for this game.
		c.gameInfoById[msg.GameId] = gameInfo{gameName: msg.GameName}
	}

	if info.votes >= c.target { // Reviews have been received, and they exceed the target vote count.
		if err := c.w.Broker.Publish(c.w.Outputs[0].Exchange, c.w.Outputs[0].Key, b, map[string]any{amqp.MessageIdHeader: message.GameNameID}); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
			return
		}

		c.gameInfoById[msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
	} else { // Reviews have been received, but they do not exceed the target vote count.
		c.gameInfoById[msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}
}
