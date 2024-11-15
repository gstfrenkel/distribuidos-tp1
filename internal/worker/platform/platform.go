package platform

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
)

var (
	headersEof = map[string]any{amqp.MessageIdHeader: uint8(message.EofMsg)}
	headers    = map[string]any{amqp.MessageIdHeader: uint8(message.PlatformID)}
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

func (f *filter) Init() error {
	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, header amqp.Header) {
	var sequenceIds []sequence.Destination

	switch header.MessageId {
	case message.EofMsg:
		headersEof[amqp.ClientIdHeader] = header.ClientId
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, headersEof, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	case message.GameIdMsg:
		msg, err := message.GameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		sequenceIds = f.publish(header, msg)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), header.MessageId)
	}

	if err := f.w.Recovery.Log(recovery.NewRecord(header, sequenceIds, []string{})); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToLog.Error(), err)
	}
}

func (f *filter) publish(header amqp.Header, msg message.Game) []sequence.Destination {
	sequenceId := f.w.NextSequenceId(f.w.Outputs[0].Key)
	headers[amqp.ClientIdHeader] = header.ClientId
	headers[amqp.SequenceIdHeader] = sequence.SrcNew(f.w.Id, sequenceId).ToString()

	platforms := msg.ToPlatformMessage()
	b, err := platforms.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	key := f.w.Outputs[0].Key
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}
