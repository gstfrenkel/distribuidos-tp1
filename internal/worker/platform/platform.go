package platform

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
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

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var err error

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds, err = f.w.HandleEofMessage(delivery.Body, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	case message.GameIdMsg:
		msg, err := message.GameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = f.publish(headers, msg)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(headers amqp.Header, msg message.Game) []sequence.Destination {
	output := f.w.Outputs[0]
	sequenceId := f.w.NextSequenceId(output.Key)
	headers = headers.WithMessageId(message.PlatformID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

	platforms := msg.ToPlatformMessage()
	b, err := platforms.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	if err = f.w.Broker.Publish(output.Exchange, output.Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return []sequence.Destination{sequence.DstNew(output.Key, sequenceId)}
}
