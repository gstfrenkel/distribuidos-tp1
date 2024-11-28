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

func (p *percentile) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var msg []byte

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = processEof(
			headers,
			p.eofsByClient,
			p.gameInfoByClient,
			p.processEof,
		)
	case message.ScoredReviewID:
		p.processReview(headers, delivery.Body, false)
	case message.GameNameID:
		p.processGame(headers, delivery.Body, false)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, msg
}

func (p *percentile) processEof(headers amqp.Header) []sequence.Destination {
	var sequenceIds []sequence.Destination

	userInfo, ok := p.gameInfoByClient[headers.ClientId]
	if ok {
		sequenceIds = p.processBatch(headers, userInfo)
	}

	auxSequenceIds, err := p.w.HandleEofMessage(amqp.EmptyEof, headers)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	sequenceIds = append(sequenceIds, auxSequenceIds...)
	return sequenceIds
}

func (p *percentile) processBatch(headers amqp.Header, userInfo map[int64]gameInfo) []sequence.Destination {
	var sequenceIds []sequence.Destination
	var reviews message.ScoredReviews

	for id, info := range userInfo {
		if info.gameName == "" || info.votes == 0 {
			continue
		}

		reviews = append(reviews, message.ScoredReview{GameId: id, Votes: info.votes, GameName: info.gameName})

		if len(reviews) >= int(p.batchSize) {
			sequenceIds = append(sequenceIds, p.publish(headers, reviews))
			reviews = reviews[:0] // Reset slice without deallocating memory.
		}
	}

	if len(reviews) > 0 {
		sequenceIds = append(sequenceIds, p.publish(headers, reviews))
	}

	return sequenceIds
}

func (p *percentile) processReview(headers amqp.Header, msgBytes []byte, _ bool) []sequence.Destination {
	msg, err := message.ScoredReviewFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return nil
	}

	userInfo, ok := p.gameInfoByClient[headers.ClientId]
	if !ok { // First message for this clientId
		p.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {votes: msg.Votes}}
		return nil
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // First message for this gameId
		p.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{votes: msg.Votes}
	} else {
		p.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}
	}

	return nil
}

func (p *percentile) processGame(headers amqp.Header, msgBytes []byte, _ bool) []sequence.Destination {
	msg, err := message.GameNameFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return nil
	}

	userInfo, ok := p.gameInfoByClient[headers.ClientId]
	if !ok { // First message for this clientId
		p.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return nil
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // First message for this gameId
		p.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName}
	} else {
		p.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}
	return nil
}

func (p *percentile) publish(headers amqp.Header, reviews message.ScoredReviews) sequence.Destination {
	b, err := reviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	key := p.w.Outputs[0].Key
	sequenceId := p.w.NextSequenceId(key)

	headers = headers.WithMessageId(message.ScoredReviewID).WithSequenceId(sequence.SrcNew(p.w.Id, sequenceId))

	if err = p.w.Broker.Publish(p.w.Outputs[0].Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequence.DstNew(key, sequenceId)
}

func (p *percentile) recover() {
	ch := make(chan recovery.Message, worker.ChanSize)
	go p.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			processEof(
				recoveredMsg.Header(),
				p.eofsByClient,
				p.gameInfoByClient,
				nil,
			)
		case message.ScoredReviewID:
			p.processReview(recoveredMsg.Header(), recoveredMsg.Message(), true)
		case message.GameNameID:
			p.processGame(recoveredMsg.Header(), recoveredMsg.Message(), true)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
