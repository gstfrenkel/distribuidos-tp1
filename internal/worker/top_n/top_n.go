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
	w   *worker.Worker
	top PriorityQueue //top n games
	n   int
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	f.n = int(f.w.Query.(float64))
	f.top = make(PriorityQueue, 0, f.n)
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

		f.updateTop(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) updateTop(msg message.ScoredReview) {
	notInTop := f.fixHeap(msg)

	if notInTop {
		if f.top.Len() < f.n {
			heap.Push(&f.top, &msg)
		} else if msg.Votes > f.top[0].Votes { //the game has more votes than the lowest in the top N
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
			f.top[i].Votes = msg.Votes
			heap.Fix(&f.top, i)
			return false
		}
	}
	return true
}

// Eof msg received, so all msgs were received too.
// Send the top n games to the broker
func (f *filter) publish() {
	headers := map[string]any{amqp.MessageIdHeader: uint8(message.ScoredReviewID)}
	topNScoredReviews := f.getTopNScoredReviews()
	bytes, err := topNScoredReviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
	logs.Logger.Debugf("Top %d games sent", f.n)
}

func (f *filter) getTopNScoredReviews() message.ScoredReviews {
	topNAsSlice := make([]message.ScoredReview, f.top.Len())
	for i := 0; i < f.n; i++ {
		topNAsSlice[(f.n-1)-i] = *heap.Pop(&f.top).(*message.ScoredReview)
	}
	return topNAsSlice
}
