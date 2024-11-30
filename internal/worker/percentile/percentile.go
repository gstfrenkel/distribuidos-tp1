package percentile

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

type filter struct {
	w             *worker.Worker
	n             uint8                            //percentile value (0-100)
	scoredReviews map[string]message.ScoredReviews // <clientid, scoredReviews>
	batchSize     uint16
	eofsRecv      map[string]uint8 // <clientid, eofsRecv>
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	params := f.w.Query.([]any)
	f.n = uint8(params[0].(float64))
	f.batchSize = uint16(params[1].(float64))
	f.scoredReviews = make(map[string]message.ScoredReviews)
	f.eofsRecv = make(map[string]uint8)

	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	headers = headers.WithOriginId(amqp.Query5originId)

	switch headers.MessageId {
	case message.EofMsg:
		f.eofsRecv[headers.ClientId]++
		if f.eofsRecv[headers.ClientId] >= f.w.ExpectedEofs {
			f.publish(headers)
		}
	case message.ScoredReviewID:
		msg, err := message.ScoredReviewsFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			f.saveScoredReview(msg, headers.ClientId)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) saveScoredReview(msg message.ScoredReviews, clientId string) {
	if _, ok := f.scoredReviews[clientId]; !ok {
		f.scoredReviews[clientId] = make(message.ScoredReviews, 0)
	}

	f.scoredReviews[clientId] = append(f.scoredReviews[clientId], msg...)
}

func (f *filter) publish(headers amqp.Header) {
	output := f.w.Outputs[0]
	output.Key = shard.String(headers.SequenceId, f.w.Outputs[0].Key, f.w.Outputs[0].Consumers)

	if games := f.getGamesInPercentile(headers.ClientId); games != nil {
		f.sendBatches(headers, output, games)
	}

	f.sendEof(headers, output)
	f.reset(headers.ClientId)
}

func (f *filter) sendBatch(bytes []byte, headers amqp.Header, output amqp.Destination) {
	headers = headers.WithMessageId(message.ScoredReviewID)

	if err := f.w.Broker.Publish(output.Exchange, output.Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Games in percentile %d published", f.n)
}

func (f *filter) getGamesInPercentile(clientId string) message.ScoredReviews {
	if reviews, ok := f.scoredReviews[clientId]; ok {
		reviews.Sort(true)
		return reviews[f.percentileIdx(clientId):]
	}
	return nil
}

func (f *filter) percentileIdx(clientId string) int {
	length := len(f.scoredReviews[clientId])
	percentileIndex := int((float64(f.n) / 100) * float64(length))
	if percentileIndex >= length {
		percentileIndex = length - 1
	}
	return percentileIndex
}

func (f *filter) reset(clientId string) {
	delete(f.scoredReviews, clientId)
	delete(f.eofsRecv, clientId)
}

func (f *filter) sendEof(headers amqp.Header, output amqp.Destination) {
	_, err := f.w.HandleEofMessage(amqp.EmptyEof, headers, amqp.DestinationEof(output)) // TODO: Return sequence IDs
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Eof message sent for client %s", headers.ClientId)
}

func (f *filter) sendBatches(headers amqp.Header, output amqp.Destination, msg message.ScoredReviews) {
	for start := 0; start < len(msg); {
		batch, nextStart := f.nextBatch(msg, start, len(msg))
		bytes, err := batch.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err)
			return
		}
		f.sendBatch(bytes, headers, output)
		start = nextStart
	}
}

func (f *filter) nextBatch(data message.ScoredReviews, start int, gamesLen int) (message.ScoredReviews, int) {
	end := start + int(f.batchSize)
	if end > gamesLen {
		end = gamesLen
	}
	return data[start:end], end
}
