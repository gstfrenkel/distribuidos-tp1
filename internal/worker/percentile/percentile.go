package percentile

import (
	"sort"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
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

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)
	if messageId == message.EofMsg {
		f.eofsRecv[clientId]++
		if f.eofsRecv[clientId] >= f.w.Peers {
			f.publish(clientId)
		}
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewsFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.saveScoredReview(msg, clientId)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) saveScoredReview(msg message.ScoredReviews, clientId string) {
	if _, ok := f.scoredReviews[clientId]; !ok {
		f.scoredReviews[clientId] = make(message.ScoredReviews, 0)
	}

	f.scoredReviews[clientId] = append(f.scoredReviews[clientId], msg...)
}

func (f *filter) publish(clientId string) {
	games := f.getGamesInPercentile(clientId)
	if games != nil {
		SendBatches(games.ToAny(), f.batchSize, message.ScoredRevFromAnyToBytes, f.sendBatch, clientId)
	}

	f.sendEof(clientId)
	f.reset(clientId)
}

func (f *filter) sendBatch(bytes []byte, clientId string) {
	headers := map[string]any{amqp.MessageIdHeader: uint8(message.ScoredReviewID), amqp.OriginIdHeader: amqp.Query5originId, amqp.ClientIdHeader: clientId}
	if err := f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
	logs.Logger.Infof("Games in percentile %d published", f.n)
}

func (f *filter) getGamesInPercentile(clientId string) message.ScoredReviews {
	if reviews, ok := f.scoredReviews[clientId]; ok {
		f.sortScoredReviews(clientId)
		return reviews[f.percentileIdx(clientId):]
	}

	return nil
}

func (f *filter) sortScoredReviews(clientId string) {
	sort.Slice(f.scoredReviews[clientId], func(i, j int) bool {
		return f.scoredReviews[clientId][i].Votes < f.scoredReviews[clientId][j].Votes
	})
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

func (f *filter) sendEof(clientId string) {
	headers := map[string]any{amqp.OriginIdHeader: amqp.Query5originId, amqp.ClientIdHeader: clientId}
	if err := f.w.Broker.HandleEofMessage(f.w.Id, 0, amqp.EmptyEof, headers, f.w.InputEof, f.w.OutputsEof...); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Eof message sent for client %s", clientId)
}
