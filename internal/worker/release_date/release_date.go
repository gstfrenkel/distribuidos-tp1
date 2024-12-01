package release_date

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
	"tp1/pkg/utils/shard"
)

type filter struct {
	w         *worker.Worker
	startYear int
	endYear   int
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}
	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *filter) Start() {
	slice := f.w.Query.([]any)
	f.startYear = int(slice[0].(float64))
	f.endYear = int(slice[1].(float64))
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
	case message.GameReleaseID:
		msg, err := message.ReleasesFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			sequenceIds = f.publish(msg, headers)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(msg message.Releases, headers amqp.Header) []sequence.Destination {
	dateFilteredGames := msg.ToPlaytimeMessage(f.startYear, f.endYear)
	output := f.w.Outputs[0]

	b, err := dateFilteredGames.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return []sequence.Destination{}
	}

	key := shard.String(headers.SequenceId, output.Key, output.Consumers)
	sequenceId := f.w.NextSequenceId(key)

	headers = headers.WithMessageId(message.GameWithPlaytimeID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

	if err = f.w.Broker.Publish(output.Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func (f *filter) recover() {
	f.w.Recover(nil)
}
