package top_n_playtime

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

type filter struct {
	w           *worker.Worker
	n           uint8
	clientHeaps map[string]*MinHeapPlaytime
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}
	return &filter{
		w:           w,
		clientHeaps: make(map[string]*MinHeapPlaytime),
	}, nil
}

func (f *filter) Init() error {
	return f.w.Init()
}

func (f *filter) Start() {
	f.n = uint8(f.w.Query.(float64))
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	headers = headers.WithOriginId(amqp.Query2originId)

	switch headers.MessageId {
	case message.EofMsg:
		workersVisited, err := message.EofFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return nil, nil
		}

		if !workersVisited.Contains(f.w.Id) {
			f.publish(headers)
			delete(f.clientHeaps, headers.ClientId)
		}

		_, err = f.w.HandleEofMessage(delivery.Body, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	case message.GameWithPlaytimeID:
		if _, exists := f.clientHeaps[headers.ClientId]; !exists {
			f.clientHeaps[headers.ClientId] = &MinHeapPlaytime{}
		}
		clientHeap := f.clientHeaps[headers.ClientId]

		msg, err := message.DateFilteredReleasesFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return nil, nil
		}
		clientHeap.UpdateReleases(msg, int(f.n))
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(headers amqp.Header) {
	clientHeap, exists := f.clientHeaps[headers.ClientId]
	if !exists || clientHeap == nil {
		return
	}

	topNPlaytime := ToTopNPlaytimeMessage(f.n, clientHeap)
	logs.Logger.Infof("Top %d games with most playtime for client %s: %v", f.n, headers.ClientId, topNPlaytime)
	b, err := topNPlaytime.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	headers = headers.WithMessageId(message.GameWithPlaytimeID)
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
