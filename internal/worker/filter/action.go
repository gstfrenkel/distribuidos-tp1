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

type action struct {
	w *worker.Worker
}

func NewAction() (worker.W, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}
	return &action{w: w}, nil
}

func (f *action) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *action) Start() {
	f.w.Start(f)
}

func (f *action) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination
	var err error

	switch headers.MessageId {
	case message.EofMsg:
		sequenceIds, err = f.w.HandleEofMessage(delivery.Body, headers.WithOriginId(amqp.GameOriginId))
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

func (f *action) publish(headers amqp.Header, msg message.Game) []sequence.Destination {
	games := msg.ToGameNamesMessage(f.w.Query.(string))
	sequenceIds := make([]sequence.Destination, 0, len(games)*len(f.w.Outputs))

	for _, game := range games {
		b, err := game.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		for _, output := range f.w.Outputs {
			key := shard.Int64(game.GameId, output.Key, output.Consumers)
			sequenceId := f.w.NextSequenceId(key)
			sequenceIds = append(sequenceIds, sequence.DstNew(key, sequenceId))

			headers = headers.WithMessageId(message.GameNameID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

			if err = f.w.Broker.Publish(output.Exchange, key, b, headers); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
			}
		}
	}

	return sequenceIds
}

func (f *action) recover() {
	f.w.Recover(nil)
}
