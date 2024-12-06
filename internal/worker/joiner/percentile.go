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
	joiner    *joiner
	batchSize uint16
}

func NewPercentile() (worker.Node, error) {
	j, err := newJoiner()
	if err != nil {
		return nil, err
	}

	return &percentile{
		joiner:    j,
		batchSize: uint16(j.w.Query.(float64)),
	}, nil
}

func (p *percentile) Init() error {
	if err := p.joiner.w.Init(); err != nil {
		return err
	}

	p.recover()

	return nil
}

func (p *percentile) Start() {
	p.joiner.w.Start(p)
}

func (p *percentile) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte, bool) {
	var sequenceIds []sequence.Destination
	var done bool

	switch headers.MessageId {
	case message.EofId:
		sequenceIds, done = p.joiner.processEof(headers, p.processEof)
	case message.ScoredReviewId:
		p.processReview(headers, delivery.Body, false)
	case message.GameNameId:
		p.processGame(headers, delivery.Body, false)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, delivery.Body, done
}

func (p *percentile) processEof(headers amqp.Header) []sequence.Destination {
	var sequenceIds []sequence.Destination

	userInfo, ok := p.joiner.gameInfoByClient[headers.ClientId]
	if ok {
		sequenceIds = p.processBatch(headers, userInfo)
	}

	auxSequenceIds, err := p.joiner.w.HandleEofMessage(amqp.EmptyEof, headers)
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

	userInfo, ok := p.joiner.gameInfoByClient[headers.ClientId]
	if !ok { // First message for this clientId
		p.joiner.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {votes: msg.Votes}}
		return nil
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // First message for this gameId
		p.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{votes: msg.Votes}
	} else {
		p.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: info.gameName, votes: info.votes + msg.Votes}
	}

	return nil
}

func (p *percentile) processGame(headers amqp.Header, msgBytes []byte, _ bool) []sequence.Destination {
	msg, err := message.GameNameFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return nil
	}

	userInfo, ok := p.joiner.gameInfoByClient[headers.ClientId]
	if !ok { // First message for this clientId
		p.joiner.gameInfoByClient[headers.ClientId] = map[int64]gameInfo{msg.GameId: {gameName: msg.GameName}}
		return nil
	}

	info, ok := userInfo[msg.GameId]
	if !ok { // First message for this gameId
		p.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName}
	} else {
		p.joiner.gameInfoByClient[headers.ClientId][msg.GameId] = gameInfo{gameName: msg.GameName, votes: info.votes}
	}
	return nil
}

func (p *percentile) publish(headers amqp.Header, reviews message.ScoredReviews) sequence.Destination {
	b, err := reviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	key := p.joiner.w.Outputs[0].Key
	sequenceId := p.joiner.w.NextSequenceId(key)

	headers = headers.WithMessageId(message.ScoredReviewId).WithSequenceId(sequence.SrcNew(p.joiner.w.Uuid, sequenceId))

	if err = p.joiner.w.Broker.Publish(p.joiner.w.Outputs[0].Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequence.DstNew(key, sequenceId)
}

func (p *percentile) recover() {
	p.joiner.recover(p)
}
