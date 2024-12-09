package filter

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

type platform struct {
	w *worker.Worker
}

func NewPlatform() (worker.Node, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}
	return &platform{w: w}, nil
}

func (f *platform) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *platform) Start() {
	f.w.Start(f)
}

func (f *platform) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var err error

	switch headers.MessageId {
	case message.EofId:
		sequenceIds, err = f.w.HandleEofMessage(delivery.Body, headers)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	case message.GameId:
		msg, err := message.GamesFromBytes(delivery.Body)
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

func (f *platform) publish(headers amqp.Header, msg message.Game) []sequence.Destination {
	output := f.w.Outputs[0]
	key := shard.String(headers.SequenceId, output.Key, output.Consumers)
	sequenceId := f.w.NextSequenceId(key)
	headers = headers.WithMessageId(message.PlatformId).WithSequenceId(sequence.SrcNew(f.w.Uuid, sequenceId))

	platforms := msg.ToPlatformMessage()
	b, err := platforms.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	if err = f.w.Broker.Publish(output.Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func (f *platform) recover() {
	f.w.Recover(nil)
}
