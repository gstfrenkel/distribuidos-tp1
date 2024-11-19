package joiner

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
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

func (p *percentile) Process(delivery amqp.Delivery, header amqp.Header) ([]sequence.Destination, []string) {
	var sequenceIds []sequence.Destination
	var recvMsg []string

	switch header.MessageId {
	case message.EofMsg:
		sequenceIds = processEof(
			header,
			p.eofsByClient,
			p.gameInfoByClient,
			p.processEof,
		)
	case message.ScoredReviewID:
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			p.processReview(header.ClientId, msg)
		}
	case message.GameNameID:
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			p.processGame(header.ClientId, msg)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), header.MessageId)
	}

	return sequenceIds, recvMsg
}

func (p *percentile) processEof(clientId string) []sequence.Destination {
	var sequenceIds []sequence.Destination

	userInfo, ok := p.gameInfoByClient[clientId]
	if ok {
		sequenceIds = p.processBatch(clientId, userInfo)
	}

	auxSequenceIds, err := p.w.HandleEofMessage(amqp.EmptyEof, headersEof)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	sequenceIds = append(sequenceIds, auxSequenceIds...)
	return sequenceIds
}

func (p *percentile) processBatch(clientId string, userInfo map[int64]gameInfo) []sequence.Destination {
	var sequenceIds []sequence.Destination
	var reviews message.ScoredReviews

	for id, info := range userInfo {
		if info.gameName == "" || info.votes == 0 {
			continue
		}

		reviews = append(reviews, message.ScoredReview{GameId: id, Votes: info.votes, GameName: info.gameName})

		if len(reviews) >= int(p.batchSize) {
			sequenceIds = append(sequenceIds, p.publish(clientId, reviews))
			reviews = reviews[:0] // Reset slice without deallocating memory.
		}
	}

	if len(reviews) > 0 {
		sequenceIds = append(sequenceIds, p.publish(clientId, reviews))
	}

	return sequenceIds
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

func (p *percentile) publish(clientId string, reviews message.ScoredReviews) sequence.Destination {
	b, err := reviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	key := p.w.Outputs[0].Key
	sequenceId := p.w.NextSequenceId(key)
	headersReview[amqp.ClientIdHeader] = clientId
	headersReview[amqp.SequenceIdHeader] = sequence.SrcNew(p.w.Id, sequenceId)
	if err = p.w.Broker.Publish(p.w.Outputs[0].Exchange, key, b, headersReview); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequence.DstNew(key, sequenceId)
}
