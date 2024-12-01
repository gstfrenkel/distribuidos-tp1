package aggregator

import (
	"math"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
)

type percentile struct {
	w             *worker.Worker
	n             uint8                            //percentile value (0-100)
	scoredReviews map[string]message.ScoredReviews // <clientid, scoredReviews>
	batchSize     uint16
	eofsRecv      map[string]uint8 // <clientid, eofsRecv>
}

func NewPercentile() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	params := w.Query.([]any)

	return &percentile{
		w:             w,
		n:             uint8(params[0].(float64)),
		batchSize:     uint16(params[1].(float64)),
		scoredReviews: make(map[string]message.ScoredReviews),
		eofsRecv:      make(map[string]uint8),
	}, nil
}

func (f *percentile) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *percentile) Start() {
	f.w.Start(f)
}

func (f *percentile) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = f.processEof(headers, false)
	case message.ScoredReviewID:
		f.saveScoredReview(delivery.Body, headers.ClientId)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *percentile) processEof(headers amqp.Header, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	f.eofsRecv[headers.ClientId]++
	if f.eofsReached(headers) {
		if !recovery {
			sequenceIds = f.publish(headers, recovery)
		}

		sequenceIds = append(sequenceIds, f.sendEof(headers)...)
		f.reset(headers.ClientId)
	}
	return sequenceIds
}

func (f *percentile) eofsReached(headers amqp.Header) bool {
	return f.eofsRecv[headers.ClientId] >= f.w.ExpectedEofs
}

func (f *percentile) saveScoredReview(msgBytes []byte, clientId string) {
	msg, err := message.ScoredReviewsFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	if _, ok := f.scoredReviews[clientId]; !ok {
		f.scoredReviews[clientId] = make(message.ScoredReviews, 0)
	}

	f.scoredReviews[clientId] = append(f.scoredReviews[clientId], msg...)
}

func (f *percentile) publish(headers amqp.Header, recovery bool) []sequence.Destination {
	output := shardOutput(f.w.Outputs[0], headers.ClientId)
	var sequenceIds []sequence.Destination

	if games := f.getGamesInPercentile(headers.ClientId); games != nil && !recovery {
		sequenceIds = f.sendBatches(headers, output, games)
	}

	return sequenceIds
}

func (f *percentile) getGamesInPercentile(clientId string) message.ScoredReviews {
	if reviews, ok := f.scoredReviews[clientId]; ok {
		reviews.Sort(true)
		return reviews[f.percentileIdx(clientId):]
	}
	return nil
}

func (f *percentile) percentileIdx(clientId string) int {
	length := len(f.scoredReviews[clientId])
	percentileIndex := int((float64(f.n) / 100) * float64(length))
	if percentileIndex >= length {
		percentileIndex = length - 1
	}
	return percentileIndex
}

func (f *percentile) reset(clientId string) {
	delete(f.scoredReviews, clientId)
	delete(f.eofsRecv, clientId)
}

// TODO extract
func (f *percentile) sendEof(headers amqp.Header) []sequence.Destination {
	output := shardOutput(f.w.Outputs[0], headers.ClientId)
	sequenceIds, err := f.w.HandleEofMessage(amqp.EmptyEof, headers.WithOriginId(amqp.Query5originId), amqp.DestinationEof(output)) // TODO: Return sequence IDs
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Debugf("Eof message sent for client %s", headers.ClientId)
	return sequenceIds
}

func (f *percentile) sendBatches(headers amqp.Header, output amqp.Destination, msg message.ScoredReviews) []sequence.Destination {
	numberOfBatches := int(math.Ceil(float64(len(msg)) / float64(f.batchSize)))
	sequenceIds := make([]sequence.Destination, 0, numberOfBatches)
	headers = headers.WithMessageId(message.ScoredReviewID).WithOriginId(amqp.Query5originId)

	for start := 0; start < len(msg); {
		batch, nextStart := f.nextBatch(msg, start, len(msg))
		bytes, err := batch.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err)
			return sequenceIds
		}

		sequenceId := f.w.NextSequenceId(output.Key)
		sequenceIds = append(sequenceIds, sequence.DstNew(output.Key, sequenceId))
		headers = headers.WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

		f.sendBatch(bytes, headers, output)
		start = nextStart
	}

	return sequenceIds
}

func (f *percentile) sendBatch(bytes []byte, headers amqp.Header, output amqp.Destination) {
	if err := f.w.Broker.Publish(output.Exchange, output.Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Games in percentile %d published", f.n)
}

func (f *percentile) nextBatch(data message.ScoredReviews, start int, gamesLen int) (message.ScoredReviews, int) {
	end := start + int(f.batchSize)
	if end > gamesLen {
		end = gamesLen
	}
	return data[start:end], end
}

func (f *percentile) recover() {
	ch := make(chan recovery.Message, worker.ChanSize)
	go f.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			f.processEof(recoveredMsg.Header(), true)
		case message.ScoredReviewID:
			f.saveScoredReview(recoveredMsg.Message(), recoveredMsg.Header().ClientId)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
