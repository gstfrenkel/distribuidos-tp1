package filter

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

const (
	query3 uint8 = iota
	query4
	query5
	nQueries
)

type review struct {
	w      *worker.Worker
	scores [nQueries]int8
}

func NewReview() (worker.W, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &review{w: w, scores: [nQueries]int8{}}, nil
}

func (f *review) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *review) Start() {
	slice := f.w.Query.([]any)
	f.scores[query3] = int8(slice[query3].(float64))
	f.scores[query4] = int8(slice[query4].(float64))
	f.scores[query5] = int8(slice[query5].(float64))

	f.w.Start(f)
}

func (f *review) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var err error

	headers = headers.WithOriginId(amqp.ReviewOriginId)

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds, err = f.w.HandleEofMessage(delivery.Body, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	case message.ReviewIdMsg:
		sequenceIds = f.publish(delivery.Body, headers)

	default:
		logs.Logger.Infof(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *review) publish(msgBytes []byte, headers amqp.Header) []sequence.Destination {
	msg, err := message.ReviewFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return []sequence.Destination{}
	}

	return append(f.publishReviewWithText(msg, headers), f.publishScoredReview(msg, headers)...)
}

func (f *review) publishScoredReview(msg message.Review, headers amqp.Header) []sequence.Destination {
	headers = headers.WithMessageId(message.ScoredReviewID)

	reviews := msg.ToScoredReviewMessage(f.scores[query3])
	sequenceIds := f.shardPublish(reviews, f.w.Outputs[query3], headers)

	if f.scores[query5] != f.scores[query3] {
		reviews = msg.ToScoredReviewMessage(f.scores[query5])
	}

	return append(sequenceIds, f.shardPublish(reviews, f.w.Outputs[query5], headers)...)
}

func (f *review) publishReviewWithText(msg message.Review, headers amqp.Header) []sequence.Destination {

	b, err := msg.ToReviewWithTextMessage(f.scores[query4]).ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	output := f.w.Outputs[query4]
	key := shard.String(headers.SequenceId, output.Key, output.Consumers)
	sequenceId := f.w.NextSequenceId(key)
	headers = headers.WithMessageId(message.ReviewWithTextID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

	if err = f.w.Broker.Publish(output.Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return []sequence.Destination{sequence.DstNew(output.Key, sequenceId)}
}

func (f *review) shardPublish(reviews message.ScoredReviews, output amqp.Destination, headers amqp.Header) []sequence.Destination {
	sequenceIds := make([]sequence.Destination, 0, len(reviews))
	for _, rv := range reviews {
		b, err := rv.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := shard.Int64(rv.GameId, output.Key, output.Consumers)
		sequenceId := f.w.NextSequenceId(k)
		sequenceIds = append(sequenceIds, sequence.DstNew(k, sequenceId))
		headers = headers.WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

		if err = f.w.Broker.Publish(output.Exchange, k, b, headers); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}

	return sequenceIds
}

func (f *review) recover() {
	f.w.Recover(nil)
}
