package top_n

import (
	"container/heap"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type GameId int64

type filter struct {
	w     *worker.Worker
	votes map[GameId]uint64
	top   PriorityQueue //top n games
	n     int
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	f.n = f.w.Query.(int)
	f.top = make(PriorityQueue, 0, f.n)
	f.votes = make(map[GameId]uint64)
	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))
	if messageId == message.EofMsg {
		f.sendTop()
	} else if messageId == message.ScoredReviewID {
		msg, err := message.ScoredReviewFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.updateTop(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) updateTop(msg message.ScoredReview) {
	f.updateVotes(msg.GameId, msg.Votes)
	notInTop := f.fixHeap(msg)

	if notInTop {
		if f.top.Len() < f.n {
			msg.Votes = f.votes[GameId(msg.GameId)]
			heap.Push(&f.top, &msg)
		} else if f.votes[GameId(msg.GameId)] > f.top[0].Votes {
			heap.Pop(&f.top)
			heap.Push(&f.top, &msg)
		}
	}
}

// If the game is already in the top, fix the heap and return false
// If the game is not in the top, return true
func (f *filter) fixHeap(msg message.ScoredReview) bool {
	for i, item := range f.top {
		if item.GameId == msg.GameId {
			f.top[i].Votes = f.votes[GameId(msg.GameId)]
			heap.Fix(&f.top, i)
			return false
		}
	}
	return true
}

func (f *filter) updateVotes(gameId int64, votes uint64) {
	if _, ok := f.votes[GameId(gameId)]; ok {
		f.votes[GameId(gameId)] += votes
	} else {
		f.votes[GameId(gameId)] = votes
	}
}

// Eof msg received, so all msgs were received too.
// Send the top n games to the broker
func (f *filter) sendTop() {
	headers := map[string]any{amqp.MessageIdHeader: message.ScoredReviewID}
	topNScoredReviews := f.getTopNScoredReviews()
	bytes, err := topNScoredReviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
	logs.Logger.Info("Top N games sent")

	f.publishEof()
}

func (f *filter) getTopNScoredReviews() message.ScoredReviews {
	topNAsSlice := make([]message.ScoredReview, f.top.Len())
	for i := 0; i < f.n; i++ {
		topNAsSlice[i] = *heap.Pop(&f.top).(*message.ScoredReview)
	}
	return topNAsSlice
}

func (f *filter) publishEof() {
	headers := map[string]any{amqp.MessageIdHeader: message.EofMsg}
	if err := f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, nil, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
	logs.Logger.Info("Top N games eof sent")
}
