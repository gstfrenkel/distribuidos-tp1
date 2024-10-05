package review

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const (
	query3 uint8 = iota
	query4
	query5
)

type filter struct {
	w      *worker.Worker
	scores []int8
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	return f.w.Init()
}

func (f *filter) Start() {
	f.scores = f.w.Query.([]int8)

	f.w.Start(f)
}

func (f *filter) Process(reviewDelivery amqpconn.Delivery) {
	messageId := message.ID(reviewDelivery.Headers[amqpconn.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, reviewDelivery.Body, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	} else if messageId == message.ReviewIdMsg {
		msg, err := message.ReviewFromBytes(reviewDelivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg)
	} else {
		logs.Logger.Infof(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(msg message.Review) {
	b, err := msg.ToReviewWithTextMessage(f.scores[query4]).ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else if err = f.w.Broker.Publish(f.w.Outputs[query4].Exchange, "", uint8(message.ReviewWithTextID), b); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	reviews := msg.ToScoredReviewMessage(f.scores[query3])
	f.shardPublish(reviews, f.w.Outputs[query3])

	if f.scores[query5] != f.scores[query3] {
		reviews = msg.ToScoredReviewMessage(f.scores[query5])
	}
	f.shardPublish(reviews, f.w.Outputs[query5])
}

func (f *filter) shardPublish(reviews message.ScoredReviews, output broker.Destination) {
	for _, rv := range reviews {
		b, err := rv.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := worker.ShardGameId(rv.GameId, output.Key, output.Consumers)
		if err = f.w.Broker.Publish(output.Exchange, k, uint8(message.ScoredReviewID), b); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}
