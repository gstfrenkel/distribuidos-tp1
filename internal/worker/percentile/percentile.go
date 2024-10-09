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
	n             uint8 //percentile value (0-100)
	scoredReviews message.ScoredReviews
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	f.n = uint8(f.w.Query.(float64))
	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	if messageId == message.EofMsg {
		f.publish()
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewsFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.saveScoredReview(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) saveScoredReview(msg message.ScoredReviews) {
	f.scoredReviews = append(f.scoredReviews, msg...)
}

func (f *filter) publish() {
	bytes, err := f.getGamesInPercentile().ToGameNameBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err)
		return
	}

	headers := map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID)}
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
	logs.Logger.Infof("Games in percentile %d published", f.n)
}

func (f *filter) getGamesInPercentile() message.ScoredReviews {
	f.sortScoredReviews()
	return f.scoredReviews[f.percentileIdx():]
}

func (f *filter) sortScoredReviews() {
	sort.Slice(f.scoredReviews, func(i, j int) bool {
		return f.scoredReviews[i].Votes < f.scoredReviews[j].Votes
	})
}

func (f *filter) percentileIdx() int {
	length := len(f.scoredReviews)
	percentileIndex := int((float64(f.n) / 100) * float64(length))
	if percentileIndex >= length {
		percentileIndex = length - 1
	}
	return percentileIndex
}
