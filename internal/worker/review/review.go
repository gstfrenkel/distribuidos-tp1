package review

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

const (
	query3 uint8 = iota
	query4
	query5
)

var (
	headersEof    = map[string]any{amqp.OriginIdHeader: amqp.ReviewOriginId, amqp.MessageIdHeader: uint8(message.EofMsg)}
	headersText   = map[string]any{amqp.MessageIdHeader: uint8(message.ReviewWithTextID)}
	headersScored = map[string]any{amqp.MessageIdHeader: uint8(message.ScoredReviewID)}
)

type filter struct {
	w      *worker.Worker
	scores [3]int8
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w, scores: [3]int8{}}, nil
}

func (f *filter) Init() error {
	return f.w.Init()
}

func (f *filter) Start() {
	slice := f.w.Query.([]any)
	f.scores[query3] = int8(slice[query3].(float64))
	f.scores[query4] = int8(slice[query4].(float64))
	f.scores[query5] = int8(slice[query5].(float64))

	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, _ amqp.Header) ([]sequence.Destination, []string) {
	var sequenceIds []sequence.Destination

	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		headersEof[amqp.ClientIdHeader] = delivery.Headers[amqp.ClientIdHeader]
		_, err := f.w.HandleEofMessage(delivery.Body, headersEof)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	} else if messageId == message.ReviewIdMsg {
		msg, err := message.ReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			headersText[amqp.ClientIdHeader] = delivery.Headers[amqp.ClientIdHeader]
			headersScored[amqp.ClientIdHeader] = delivery.Headers[amqp.ClientIdHeader]
			f.publish(msg)
		}
	} else {
		logs.Logger.Infof(errors.InvalidMessageId.Error(), messageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(msg message.Review) {
	b, err := msg.ToReviewWithTextMessage(f.scores[query4]).ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else if err = f.w.Broker.Publish(f.w.Outputs[query4].Exchange, "", b, headersText); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	reviews := msg.ToScoredReviewMessage(f.scores[query3])
	f.shardPublish(reviews, f.w.Outputs[query3])

	if f.scores[query5] != f.scores[query3] {
		reviews = msg.ToScoredReviewMessage(f.scores[query5])
	}
	f.shardPublish(reviews, f.w.Outputs[query5])
}

func (f *filter) shardPublish(reviews message.ScoredReviews, output amqp.Destination) {
	for _, rv := range reviews {
		b, err := rv.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := worker.ShardGameId(rv.GameId, output.Key, output.Consumers)
		if err = f.w.Broker.Publish(output.Exchange, k, b, headersScored); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}
