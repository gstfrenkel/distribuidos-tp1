package top_n_playtime

import (
	"strings"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

type filter struct {
	w           *worker.Worker
	n           uint8
	clientHeaps map[string]*MinHeapPlaytime
	agg         bool
}

func New() (worker.Node, error) {
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
	if err := f.w.Init(); err != nil {
		return err
	}

	f.n = uint8(f.w.Query.(float64))
	f.agg = strings.Contains(f.w.Outputs[0].Key, "%d")

	f.recover()

	return nil
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	switch headers.MessageId {
	case message.EofId:
		sequenceIds = f.processEof(delivery.Body, headers, false)
	case message.GameWithPlaytimeId:
		f.processGame(delivery.Body, headers.ClientId)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, delivery.Body
}

func (f *filter) processGame(msgBytes []byte, clientId string) {
	if _, exists := f.clientHeaps[clientId]; !exists {
		f.clientHeaps[clientId] = &MinHeapPlaytime{}
	}
	clientHeap := f.clientHeaps[clientId]

	msg, err := message.DateFilteredReleasesFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	clientHeap.UpdateReleases(msg, int(f.n))
}

func (f *filter) processEof(msgBytes []byte, headers amqp.Header, recovery bool) []sequence.Destination {
	var sequenceIds []sequence.Destination
	headers = headers.WithOriginId(amqp.Query2OriginId)

	workersVisited, err := message.EofFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return sequenceIds
	}

	if !workersVisited.Contains(f.w.Id) {
		if !recovery {
			sequenceIds = f.publish(headers)
		}
		delete(f.clientHeaps, headers.ClientId)
	}

	if !f.agg && !recovery {
		eofSqIds, err := f.w.HandleEofMessage(msgBytes, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		} else {
			sequenceIds = append(sequenceIds, eofSqIds...)
		}
	}

	return sequenceIds
}

func (f *filter) publish(headers amqp.Header) []sequence.Destination {
	var sequenceIds []sequence.Destination

	clientHeap, exists := f.clientHeaps[headers.ClientId]
	if !exists || clientHeap == nil {
		return sequenceIds
	}

	topNPlaytime := ToTopNPlaytimeMessage(f.n, clientHeap)
	b, err := topNPlaytime.ToBytes()
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
	headers = headers.WithMessageId(message.GameWithPlaytimeId).WithSequenceId(sequence.SrcNew(f.w.Uuid, sequenceId))
	sequenceIds = append(sequenceIds, sequence.DstNew(output.Key, sequenceId))

	if err = f.w.Broker.Publish(output.Exchange, output.Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return sequenceIds
}

func (f *filter) recover() {
	ch := make(chan recovery.Message, worker.ChanSize)
	go f.w.Recover(ch)

	for recoveredMsg := range ch {
		switch recoveredMsg.Header().MessageId {
		case message.EofId:
			f.processEof(recoveredMsg.Message(), recoveredMsg.Header(), true)
		case message.GameWithPlaytimeId:
			f.processGame(recoveredMsg.Message(), recoveredMsg.Header().ClientId)
		default:
			logs.Logger.Errorf(errors.InvalidMessageId.Error(), recoveredMsg.Header().MessageId)
		}
	}
}
