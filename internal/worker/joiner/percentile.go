package joiner

import (
	"fmt"

	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type percentile struct {
	w             *worker.Worker
	recvReviewEof bool
	recvGameEof   bool
	gameInfoById  map[int64]gameInfo
	batchSize     uint16
}

func NewPercentile() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &percentile{w: w, gameInfoById: map[int64]gameInfo{}}, nil
}

func (p *percentile) Init() error {
	return p.w.Init()
}

func (p *percentile) Start() {
	p.batchSize = uint16(p.w.Query.(float64))

	p.w.Start(p)
}

func (p *percentile) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		headersEof[amqp.ClientIdHeader] = delivery.Headers[amqp.ClientIdHeader]
		p.processEof(delivery.Headers[amqp.OriginIdHeader].(uint8))
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		p.processReview(msg)
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		p.processGame(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (p *percentile) processEof(origin uint8) {
	if origin == amqp.GameOriginId {
		p.recvGameEof = true
	} else if origin == amqp.ReviewOriginId {
		p.recvReviewEof = true
	} else {
		logs.Logger.Errorf(fmt.Sprintf("Unknown message origin ID received: %d", origin))
	}

	if !p.recvReviewEof || !p.recvGameEof {
		return
	}

	var reviews message.ScoredReviews
	for id, info := range p.gameInfoById {
		if info.gameName == "" {
			continue
		}

		reviews = append(reviews, message.ScoredReview{GameId: id, Votes: info.votes, GameName: info.gameName})

		if len(reviews) >= int(p.batchSize) {
			p.publish(reviews)
			reviews = reviews[:0] // Reset slice without deallocating memory.
		}
	}

	if len(reviews) > 0 {
		p.publish(reviews)
	}

	if err := p.w.Broker.HandleEofMessage(p.w.Id, 0, amqp.EmptyEof, nil, p.w.InputEof, p.w.OutputsEof...); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	p.recvReviewEof = false
	p.recvGameEof = false
	p.gameInfoById = map[int64]gameInfo{}
}

func (p *percentile) processReview(msg message.ScoredReview) {
	info, ok := p.gameInfoById[msg.GameId]
	if !ok {
		p.gameInfoById[msg.GameId] = gameInfo{votes: msg.Votes}
	} else {
		p.gameInfoById[msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}
	}
}

func (p *percentile) processGame(msg message.GameName) {
	info, ok := p.gameInfoById[msg.GameId]
	if !ok { // No reviews have been received for this game.
		p.gameInfoById[msg.GameId] = gameInfo{gameName: msg.GameName}
	} else {
		p.gameInfoById[msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}
}

func (p *percentile) publish(reviews message.ScoredReviews) {
	b, err := reviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else if err = p.w.Broker.Publish(p.w.Outputs[0].Exchange, p.w.Outputs[0].Key, b, map[string]any{amqp.MessageIdHeader: uint8(message.ScoredReviewID)}); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}
