package top_n

import (
	"container/heap"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

type GameId int64

type filter struct {
	w        *worker.Worker
	top      map[string]PriorityQueue //<client id, top n games>
	n        int
	eofsRecv map[string]uint8 //<client id, eofs received>
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{
			w:        w,
			top:      make(map[string]PriorityQueue),
			eofsRecv: make(map[string]uint8),
		},
		nil
}

func (f *filter) Init() error {
	f.n = int(f.w.Query.(float64))
	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	switch headers.MessageId {
	case message.EofMsg:
		f.eofsRecv[headers.ClientId]++
		if f.eofsRecv[headers.ClientId] >= f.w.Peers {
			f.publish(headers)
		}
	case message.ScoredReviewID:
		msg, err := message.ScoredReviewsFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			f.updateTop(msg, headers.ClientId)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) updateTop(msg message.ScoredReviews, clientId string) {
	for _, msg := range msg {
		notInTop := f.fixHeap(msg, clientId)
		clientTop, _ := f.top[clientId]
		if notInTop {
			if clientTop.Len() < f.n {
				heap.Push(&clientTop, &msg)
			} else if msg.Votes > clientTop[0].Votes { //the game has more votes than the lowest in the top N
				heap.Pop(&clientTop)
				heap.Push(&clientTop, &msg)
			}
			f.top[clientId] = clientTop
		}
	}
}

// If the game is already in the top, fix the heap and return false
// If the game is not in the top, return true
func (f *filter) fixHeap(msg message.ScoredReview, clientId string) bool {
	clientTop, ok := f.top[clientId]
	if !ok {
		f.top[clientId] = make(PriorityQueue, 0, f.n)
		return true
	}

	for i, item := range clientTop {
		if item.GameId == msg.GameId {
			clientTop[i].Votes = msg.Votes
			heap.Fix(&clientTop, i)
			f.top[clientId] = clientTop
			return false
		}
	}
	return true
}

// Eof msg received, so all msgs were received too.
// Send the top n games to the broker
func (f *filter) publish(headers amqp.Header) {
	headers = headers.WithMessageId(message.ScoredReviewID).WithOriginId(amqp.Query3originId)

	topNScoredReviews := f.getTopNScoredReviews(headers.ClientId)
	logs.Logger.Infof("Top %d games with most votes: %v", f.n, topNScoredReviews)

	bytes, err := topNScoredReviews.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, bytes, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	delete(f.top, headers.ClientId)

	if f.w.Peers == 0 { //it is not an aggregator
		_, err = f.w.HandleEofMessage(amqp.EmptyEof, headers, amqp.DestinationEof(f.w.Outputs[0]))
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}

	delete(f.eofsRecv, headers.ClientId)
}

func (f *filter) getTopNScoredReviews(clientId string) message.ScoredReviews {
	clientTop, ok := f.top[clientId]
	if !ok {
		logs.Logger.Errorf("Client %s has no top games", clientId)
		return make(message.ScoredReviews, 0)
	}

	length := clientTop.Len()
	topNAsSlice := make(message.ScoredReviews, length)
	for i := 0; i < length; i++ {
		topNAsSlice[(length-1)-i] = *heap.Pop(&clientTop).(*message.ScoredReview)
	}
	return topNAsSlice
}
