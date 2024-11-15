package joiner

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type counter struct {
	w                *worker.Worker
	eofsByClient     map[string]recvEofs
	gameInfoByClient map[string]map[int64]gameInfo
	target           uint64
}

func NewCounter() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &counter{
		w:                w,
		eofsByClient:     map[string]recvEofs{},
		gameInfoByClient: map[string]map[int64]gameInfo{},
	}, nil
}

func (c *counter) Init() error {
	c.target = uint64(c.w.Query.(float64))

	return c.w.Init()
}

func (c *counter) Start() {
	c.w.Start(c)
}

func (c *counter) Process(delivery amqp.Delivery, _ amqp.Header) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)

	if messageId == message.EofMsg {
		processEof(
			clientId,
			delivery.Headers[amqp.OriginIdHeader].(uint8),
			c.eofsByClient,
			c.gameInfoByClient,
			c.processEof,
		)
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		c.processReview(clientId, msg)
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		c.processGame(clientId, msg, delivery.Body)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (c *counter) processEof(clientId string) {
	if err := c.w.Broker.HandleEofMessage(c.w.Id, 0, amqp.EmptyEof, headersEof, c.w.InputEof, c.w.OutputsEof...); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}

func (c *counter) processReview(clientId string, msg message.ScoredReview) {
	userInfo, ok := c.gameInfoByClient[clientId]
	if !ok {
		c.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return
	}

	info, ok := userInfo[msg.GameId]
	if !ok {
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{votes: msg.Votes}
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

		headersGame[amqp.ClientIdHeader] = clientId
		if err = c.w.Broker.Publish(c.w.Outputs[0].Exchange, c.w.Outputs[0].Key, b, headersGame); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		} else {
			c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
		}
	} else {
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}
	}
}

func (c *counter) processGame(clientId string, msg message.GameName, b []byte) {
	userInfo, ok := c.gameInfoByClient[clientId]
	if !ok {
		c.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // No reviews have been received for this game.
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName}
		return
	}

	if info.votes >= c.target { // Reviews have been received, and they exceed the target vote count.
		headersGame[amqp.ClientIdHeader] = clientId
		if err := c.w.Broker.Publish(c.w.Outputs[0].Exchange, c.w.Outputs[0].Key, b, headersGame); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
			return
		}

		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
	} else { // Reviews have been received, but they do not exceed the target vote count.
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}
}
