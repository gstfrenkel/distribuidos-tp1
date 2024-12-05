package aggregator

import (
	"math"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

type percentile struct {
	agg           *aggregator
	n             uint8                            //percentile value (0-100)
	scoredReviews map[string]message.ScoredReviews // <clientid, scoredReviews>
}

func NewPercentile() (worker.Node, error) {
	a, err := newAggregator(amqp.Query5OriginId)
	if err != nil {
		return nil, err
	}

	params := a.w.Query.([]any)
	a.batchSize = uint16(params[1].(float64))

	return &percentile{
		agg:           a,
		n:             uint8(params[0].(float64)),
		scoredReviews: make(map[string]message.ScoredReviews),
	}, nil
}

func (p *percentile) Init() error {
	if err := p.agg.w.Init(); err != nil {
		return err
	}

	p.agg.recover(p, message.ScoredReviewId)

	return nil
}

func (p *percentile) Start() {
	p.agg.w.Start(p)
}

func (p *percentile) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	switch headers.MessageId {
	case message.EofId:
		sequenceIds = p.agg.processEof(p, headers.WithOriginId(p.agg.originId), false)
	case message.ScoredReviewId:
		p.save(delivery.Body, headers.ClientId)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, delivery.Body
}

func (p *percentile) save(msgBytes []byte, clientId string) {
	msg, err := message.ScoredReviewsFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	if _, ok := p.scoredReviews[clientId]; !ok {
		p.scoredReviews[clientId] = make(message.ScoredReviews, 0)
	}

	p.scoredReviews[clientId] = append(p.scoredReviews[clientId], msg...)
}

func (p *percentile) publish(headers amqp.Header) []sequence.Destination {
	output := shardOutput(p.agg.w.Outputs[0], headers.ClientId)
	var sequenceIds []sequence.Destination

	if games := p.getGamesInPercentile(headers.ClientId); games != nil {
		sequenceIds = p.sendBatches(headers, output, games)
	}

	return sequenceIds
}

func (p *percentile) getGamesInPercentile(clientId string) message.ScoredReviews {
	if reviews, ok := p.scoredReviews[clientId]; ok {
		reviews.Sort(true)
		return reviews[p.percentileIdx(clientId):]
	}
	return nil
}

func (p *percentile) percentileIdx(clientId string) int {
	length := len(p.scoredReviews[clientId])
	percentileIndex := int((float64(p.n) / 100) * float64(length))
	if percentileIndex >= length {
		percentileIndex = length - 1
	}
	return percentileIndex
}

func (p *percentile) reset(clientId string) {
	delete(p.scoredReviews, clientId)
}

func (p *percentile) sendBatches(headers amqp.Header, output amqp.Destination, msg message.ScoredReviews) []sequence.Destination {
	numberOfBatches := int(math.Ceil(float64(len(msg)) / float64(p.agg.batchSize)))
	sequenceIds := make([]sequence.Destination, 0, numberOfBatches)
	headers = headers.WithMessageId(message.ScoredReviewId).WithOriginId(p.agg.originId)

	for start := 0; start < len(msg); {
		batch, nextStart := p.nextBatch(msg, start, len(msg))
		bytes, err := batch.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err)
			return sequenceIds
		}

		sequenceId := p.agg.w.NextSequenceId(output.Key)
		sequenceIds = append(sequenceIds, sequence.DstNew(output.Key, sequenceId))
		headers = headers.WithSequenceId(sequence.SrcNew(p.agg.w.Uuid, sequenceId))

		p.sendBatch(bytes, headers, output)
		start = nextStart
	}

	return sequenceIds
}

func (p *percentile) sendBatch(bytes []byte, headers amqp.Header, output amqp.Destination) {
	if err := p.agg.w.Broker.Publish(output.Exchange, output.Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Games in percentile %d published", p.n)
}

func (p *percentile) nextBatch(data message.ScoredReviews, start int, gamesLen int) (message.ScoredReviews, int) {
	end := start + int(p.agg.batchSize)
	if end > gamesLen {
		end = gamesLen
	}
	return data[start:end], end
}
