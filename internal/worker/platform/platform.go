package platform

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
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

func (f *filter) Process(delivery amqp.Delivery, header amqp.Header) ([]sequence.Destination, []string) {
	var sequenceIds []sequence.Destination
	var err error

	switch header.MessageId {
	case message.EofMsg:
		headersEof[amqp.ClientIdHeader] = header.ClientId
		sequenceIds, err = f.w.HandleEofMessage(delivery.Body, headersEof)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	case message.GameIdMsg:
		msg, err := message.GameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = f.publish(header, msg)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), header.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(header amqp.Header, msg message.Game) []sequence.Destination {
	output := f.w.Outputs[0]

	sequenceId := f.w.NextSequenceId(output.Key)
	headers[amqp.ClientIdHeader] = header.ClientId
	headers[amqp.SequenceIdHeader] = sequence.SrcNew(f.w.Id, sequenceId).ToString()

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
