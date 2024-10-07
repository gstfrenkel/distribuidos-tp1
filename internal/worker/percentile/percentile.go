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
	w     *worker.Worker
	n     uint8 //percentile value (0-100)
	games map[uint8][]int
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	f.n = f.w.Query.(uint8)
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
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.saveScoredReview(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
	panic("implement me")
}

func (f *filter) saveScoredReview(msg message.ScoredReview) {

}

func (f *filter) publish() {

}

// calculatePercentile calculates the nth percentile of a slice of integers.
func calculatePercentile(data []int, percentile float64) int {
	if len(data) == 0 {
		return 0
	}

	sort.Ints(data)
	index := int(float64(len(data)-1) * percentile / 100.0)
	return data[index]
}
