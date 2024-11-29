package joiner

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

type counter struct {
	joiner *joiner
	target uint64
}

func NewCounter() (worker.Filter, error) {
	j, err := newJoiner()
	if err != nil {
		return nil, err
	}

	return &counter{
		joiner: j,
		target: uint64(j.w.Query.(float64)),
	}, nil
}

func (c *counter) Init() error {
	return c.joiner.w.Init()
}

func (c *counter) Start() {
	c.joiner.w.Start(c)
}

func (c *counter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var msg []byte

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = processEof(
			headers,
			c.joiner.eofsByClient,
			c.joiner.gameInfoByClient,
			c.processEof,
		)
	case message.ScoredReviewID:
		sequenceIds = c.processReview(headers, delivery.Body, false)
	case message.GameNameID:
		sequenceIds = c.processGame(headers, delivery.Body, false)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, msg
}

func (c *counter) processEof(headers amqp.Header) []sequence.Destination {
	sequenceIds, err := c.joiner.w.HandleEofMessage(amqp.EmptyEof, headers)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
	return sequenceIds
}

func (c *counter) processReview(headers amqp.Header, msgBytes []byte, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	msg, err := message.ScoredReviewFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return sequenceIds
	}

	userInfo, ok := c.joiner.gameInfoByClient[headers.ClientId]
	if !ok {
		c.joiner.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok {
		c.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{votes: msg.Votes}
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
			if sequenceIds = c.publish(headers, b); sequenceIds == nil {
				return sequenceIds
			}
		}
		c.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
	} else {
		c.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}
	}

	return sequenceIds
}

func (c *counter) processGame(headers amqp.Header, msgBytes []byte, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	msg, err := message.GameNameFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return sequenceIds
	}

	userInfo, ok := c.joiner.gameInfoByClient[headers.ClientId]
	if !ok {
		c.joiner.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return sequenceIds
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // No reviews have been received for this game.
		c.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName}
		return sequenceIds
	}

	if info.votes >= c.target {
		if !recovery {
			if sequenceIds = c.publish(headers, msgBytes); sequenceIds == nil {
				return sequenceIds
			}
		}
		c.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes, sent: true}
	} else {
		c.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}

	return sequenceIds
}

// publish publishes the game name when the target is reached.
// In case of error, returns nil
// Otherwise, returns the sequenceId of the message sent.
func (c *counter) publish(headers amqp.Header, msgBytes []byte) []sequence.Destination {
	key := c.joiner.w.Outputs[0].Key
	sequenceId := c.joiner.w.NextSequenceId(key)

	headers = headers.WithMessageId(message.GameNameID).WithSequenceId(sequence.SrcNew(c.joiner.w.Id, sequenceId))

	if err := c.joiner.w.Broker.Publish(c.joiner.w.Outputs[0].Exchange, c.joiner.w.Outputs[0].Key, msgBytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		return nil
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func (c *counter) recover() {
	c.joiner.recover(c)
}
