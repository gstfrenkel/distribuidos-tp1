package review

import (
	"fmt"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

// TODO esto se puede volar?
var (
	positiveConsumers = 1
	negativeConsumers = 1
	positiveKey       = "p%d"
	negativeKey       = "n%d"
)

type filter struct {
	w *worker.Worker
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f filter) Init() error {
	return f.w.Init()
}

func (f filter) Start() {
	f.w.Start(f)
}

func (f filter) Process(reviewDelivery amqpconn.Delivery) {
	messageId := message.ID(reviewDelivery.Headers[amqpconn.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, reviewDelivery.Body, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("\n%s\n", errors.FailedToPublish.Error())
		}
	} else if messageId == message.ReviewIdMsg {
		msg, err := message.ReviewFromBytes(reviewDelivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s\n", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg)
	} else {
		logs.Logger.Infof(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f filter) publish(msg message.Review) {
	b, err := msg.ToPositiveReviewWithTextMessage().ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s\n", errors.FailedToParse.Error(), err.Error())
	} else if err = f.w.Broker.Publish(f.w.Outputs, "", uint8(message.PositiveReviewWithTextID), b); err != nil {
		logs.Logger.Errorf("%s: %s\n", errors.FailedToPublish.Error(), err.Error())
	}

	f.shardPublish(msg.ToPositiveReviewMessage(), positiveKey, positiveConsumers, uint8(message.PositiveReviewID))
	f.shardPublish(msg.ToNegativeReviewMessage(), negativeKey, negativeConsumers, uint8(message.NegativeReviewID))
}

func (f filter) shardPublish(reviews message.ScoredReviews, k string, consumers int, id uint8) {
	for _, rv := range reviews {
		b, err := rv.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s\n", errors.FailedToParse.Error(), err.Error())
			continue
		}

		key := fmt.Sprintf(k, rv.GameId%int64(consumers))
		if err = f.w.Broker.Publish(f.w.Outputs, key, id, b); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
		logs.Logger.Infof("Published message %d to %s with key %s", id, f.w.Outputs, key)
	}
}
