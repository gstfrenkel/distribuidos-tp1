package top_n

import (
	"container/heap"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

type GameId int64

type filter struct {
	w        *worker.Worker
	top      map[string]PriorityQueue //<client id, top n games>
	n        int
	eofsRecv map[string]uint8 //<client id, eofs received>
	agg      bool
}

func New() (worker.W, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{
			w:        w,
			top:      make(map[string]PriorityQueue),
			eofsRecv: make(map[string]uint8),
			n:        int(w.Query.(float64)),
		},
		nil
}

func (f *filter) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *filter) Start() {
	f.agg = f.w.ExpectedEofs > 0

	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds = f.processEof(headers, false)
	case message.ScoredReviewID:
		f.updateTop(delivery.Body, headers.ClientId)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) processEof(headers amqp.Header, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	f.eofsRecv[headers.ClientId]++
	if f.eofsRecv[headers.ClientId] >= f.w.ExpectedEofs {
		sequenceIds = f.publish(headers, recovery)
	}

	return sequenceIds
}

func (f *filter) updateTop(msgBytes []byte, clientId string) {
	messages, err := message.ScoredReviewsFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	for _, msg := range messages {
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
func (f *filter) publish(headers amqp.Header, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	headers = headers.WithOriginId(amqp.Query3originId)

	if !recovery {
		topNScoredReviews := f.getTopNScoredReviews(headers.ClientId)
		logs.Logger.Infof("Top %d games with most votes: %v", f.n, topNScoredReviews)
		bytes, err := topNScoredReviews.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return sequenceIds
		}

		output := f.w.Outputs[0]
		if f.agg {
			output, err = shard.AggregatorOutput(output, headers.ClientId)
			if err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			}
		}

		sequenceId := f.w.NextSequenceId(output.Key)
		headers = headers.WithMessageId(message.ScoredReviewID).WithSequenceId(sequence.SrcNew(f.w.Uuid, sequenceId))
		sequenceIds = append(sequenceIds, sequence.DstNew(output.Key, sequenceId))

		if err = f.w.Broker.Publish(output.Exchange, output.Key, bytes, headers); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}

		if !f.agg {
			eofSqIds, err := f.w.HandleEofMessage(amqp.EmptyEof, headers, amqp.DestinationEof(f.w.Outputs[0]))
			if err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
			} else {
				sequenceIds = append(sequenceIds, eofSqIds...)
			}
		}
	}

	delete(f.top, headers.ClientId)
	delete(f.eofsRecv, headers.ClientId)
	return sequenceIds
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

func (f *filter) recover() {
	ch := make(chan recovery.Message, worker.ChanSize)
	go f.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofMsg:
			f.processEof(recoveredMsg.Header(), true)
		case message.ScoredReviewID:
			f.updateTop(recoveredMsg.Message(), recoveredMsg.Header().ClientId)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
