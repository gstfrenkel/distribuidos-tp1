package top_n_playtime

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
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
	f.clientHeaps = make(map[string]*MinHeapPlaytime)
	return f.w.Init()
}

func (f *filter) Start() {
	f.n = uint8(f.w.Query.(float64))
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery) {
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		workersVisited, err := message.EofFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		if !workersVisited.Contains(f.w.Id) {
			f.publish(clientId)
			delete(f.clientHeaps, clientId)
		}

		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, nil, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}

	} else if messageId == message.GameWithPlaytimeID {
		if _, exists := f.clientHeaps[clientId]; !exists {
			f.clientHeaps[clientId] = &MinHeapPlaytime{}
		}
		clientHeap := f.clientHeaps[clientId]

		msg, err := message.DateFilteredReleasesFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}
		clientHeap.UpdateReleases(msg, int(f.n))
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(clientId string) {
	clientHeap, exists := f.clientHeaps[clientId]
	if !exists || clientHeap == nil {
		return
	}

	topNPlaytime := ToTopNPlaytimeMessage(f.n, clientHeap)
	logs.Logger.Infof("Top %d games with most playtime for client %s: %v", f.n, clientId, topNPlaytime)
	b, err := topNPlaytime.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	headers := map[string]any{
		amqp.MessageIdHeader: uint8(message.GameWithPlaytimeID),
		amqp.OriginIdHeader:  amqp.Query2originId,
		amqp.ClientIdHeader:  clientId,
	}
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
