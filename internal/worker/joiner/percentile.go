package joiner

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type percentile struct {
	w                *worker.Worker
	eofsByClient     map[string]recvEofs
	gameInfoByClient map[string]map[int64]gameInfo
	batchSize        uint16
}

func NewPercentile() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &percentile{
		w:                w,
		eofsByClient:     map[string]recvEofs{},
		gameInfoByClient: map[string]map[int64]gameInfo{},
	}, nil
}

func (p *percentile) Init() error {
	p.batchSize = uint16(p.w.Query.(float64))

	return p.w.Init()
}

func (p *percentile) Start() {
	p.w.Start(p)
}

func (p *percentile) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)

	if messageId == message.EofMsg {
		processEof(
			clientId,
			delivery.Headers[amqp.OriginIdHeader].(uint8),
			p.eofsByClient,
			p.gameInfoByClient,
			p.processEof,
		)
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		p.processReview(clientId, msg)
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		p.processGame(clientId, msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (p *percentile) processEof(clientId string) {
	userInfo, ok := p.gameInfoByClient[clientId]
	if ok {
		p.processBatch(clientId, userInfo)
	}

	if err := p.w.Broker.HandleEofMessage(p.w.Id, 0, amqp.EmptyEof, headersEof, p.w.InputEof, p.w.OutputsEof...); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}

func (p *percentile) processBatch(clientId string, userInfo map[int64]gameInfo) {
	var reviews message.ScoredReviews

	for id, info := range userInfo {
		if info.gameName == "" {
			continue
		}

		reviews = append(reviews, message.ScoredReview{GameId: id, Votes: info.votes, GameName: info.gameName})

		if len(reviews) >= int(p.batchSize) {
			p.publish(clientId, reviews)
			reviews = reviews[:0] // Reset slice without deallocating memory.
		}
	}

	if len(reviews) > 0 {
		p.publish(clientId, reviews)
	}
}

func (p *percentile) processReview(clientId string, msg message.ScoredReview) {
	userInfo, ok := p.gameInfoByClient[clientId]
	if !ok { // First message for this clientId
		p.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {votes: msg.Votes}}
		return
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // First message for this gameId
		p.gameInfoByClient[clientId][msg.GameId] = gameInfo{votes: msg.Votes}
	} else {
		p.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}
	}
}

func (p *percentile) processGame(clientId string, msg message.GameName) {
	userInfo, ok := p.gameInfoByClient[clientId]
	if !ok { // First message for this clientId
		p.gameInfoByClient[clientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // First message for this gameId
		p.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName}
	} else {
		p.gameInfoByClient[clientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}
}

func (p *percentile) publish(clientId string, reviews message.ScoredReviews) {
	b, err := reviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	headersReview[amqp.ClientIdHeader] = clientId
	if err = p.w.Broker.Publish(p.w.Outputs[0].Exchange, p.w.Outputs[0].Key, b, headersReview); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}
}
