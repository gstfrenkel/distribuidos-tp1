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

func (c *counter) Process(delivery amqp.Delivery, header amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var msg []byte

	switch header.MessageId {
	case message.EofMsg:
		sequenceIds = processEof(
			header,
			c.eofsByClient,
			c.gameInfoByClient,
			c.processEof,
		)
	case message.ScoredReviewID:
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = c.processReview(header.ClientId, msg, false)
		}
	case message.GameNameID:
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = c.processGame(header.ClientId, msg, delivery.Body, false)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), header.MessageId)
	}

	return sequenceIds, msg
}

func (c *counter) processEof(_ string) []sequence.Destination {
	sequenceIds, err := c.w.HandleEofMessage(amqp.EmptyEof, headersEof)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
	return sequenceIds
}

func (c *counter) processReview(clientId string, msg message.ScoredReview, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination

	userInfo, ok := c.gameInfoByClient[clientId]
	if !ok {
		c.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok {
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{votes: msg.Votes}
		return sequenceIds
	} else if info.sent {
		return sequenceIds
	}

	if info.gameName != "" && info.votes+msg.Votes >= c.target {
		if !recovery {
			b, err := message.GameName{GameId: msg.GameId, GameName: info.gameName}.ToBytes()
			if err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
				return sequenceIds
			}

			key := c.w.Outputs[0].Key
			sequenceId := c.w.NextSequenceId(key)
			sequenceIds = append(sequenceIds, sequence.DstNew(key, sequenceId))

			headersGame[amqp.SequenceIdHeader] = sequence.SrcNew(c.w.Id, sequenceId)
			headersGame[amqp.ClientIdHeader] = clientId

			if err = c.w.Broker.Publish(c.w.Outputs[0].Exchange, c.w.Outputs[0].Key, b, headersGame); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
			}
		}
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
	} else {
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}
	}

	return sequenceIds
}

func (c *counter) processGame(clientId string, msg message.GameName, b []byte, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination

	userInfo, ok := c.gameInfoByClient[clientId]
	if !ok {
		c.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // No reviews have been received for this game.
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName}
		return sequenceIds
	}

	if info.votes >= c.target {
		if !recovery {
			key := c.w.Outputs[0].Key
			sequenceId := c.w.NextSequenceId(key)
			sequenceIds = append(sequenceIds, sequence.DstNew(key, sequenceId))

			headersGame[amqp.SequenceIdHeader] = sequence.SrcNew(c.w.Id, sequenceId)
			headersGame[amqp.ClientIdHeader] = clientId

			if err := c.w.Broker.Publish(c.w.Outputs[0].Exchange, c.w.Outputs[0].Key, b, headersGame); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
				return sequenceIds
			}
		}
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
	} else {
		c.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}

	return sequenceIds
}

func (c *counter) recover() {
	ch := make(chan recovery.Message, 32)
	go c.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			processEof(
				recoveredMsg.Header(),
				c.eofsByClient,
				c.gameInfoByClient,
				nil,
			)
		case message.ScoredReviewID:
			msg, err := message.ScoredReviewFromBytes(recoveredMsg.Message())
			if err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			} else {
				_ = c.processReview(recoveredMsg.Header().ClientId, msg, true)
			}
		case message.GameNameID:
			msg, err := message.GameNameFromBytes(recoveredMsg.Message())
			if err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			} else {
				_ = c.processGame(recoveredMsg.Header().ClientId, msg, nil, true)
			}
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
