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

const (
	q2 uint8 = iota
	q3
)

type indie struct {
	w *worker.Worker
}

func NewIndie() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &indie{w: w}, nil
}

func (f *indie) Init() error {
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *indie) Start() {
	f.w.Start(f)
}

func (f *indie) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
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

func (f *indie) publish(headers amqp.Header, msg message.Game) []sequence.Destination {
	genre := getGenre(f)
	sequenceIdsRelease := f.publishGameReleases(headers, msg, genre)
	sequenceIdsNames := f.publishGameNames(headers, msg, genre)
	return append(sequenceIdsRelease, sequenceIdsNames...)
}

func (f *indie) publishGameNames(headers amqp.Header, msg message.Game, genre string) []sequence.Destination {
	gameNames := msg.ToGameNamesMessage(genre)
	sequenceIdsNames := make([]sequence.Destination, 0, len(gameNames))
	headers = headers.WithMessageId(message.GameNameID)
	output := f.w.Outputs[q3]

	for _, game := range gameNames {
		b, err := game.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		output = f.w.Outputs[q3]
		key := shard.Int64(game.GameId, output.Key, output.Consumers)
		sequenceId := f.w.NextSequenceId(key)
		sequenceIdsNames = append(sequenceIdsNames, sequence.DstNew(key, sequenceId))

		if err = f.w.Broker.Publish(output.Exchange, key, b, headers.WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}

	return sequenceIdsNames
}

func (f *indie) publishGameReleases(headers amqp.Header, msg message.Game, genre string) []sequence.Destination {
	gameReleases := msg.ToGameReleasesMessage(genre)
	output := f.w.Outputs[q2]

	b, err := gameReleases.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return []sequence.Destination{}
	}

	key := shard.String(headers.SequenceId, output.Key, output.Consumers)
	sequenceId := f.w.NextSequenceId(key)
	headers = headers.WithMessageId(message.GameReleaseID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

	if err = f.w.Broker.Publish(output.Exchange, key, b, headers.WithMessageId(message.GameReleaseID)); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func getGenre(f *indie) string {
	return f.w.Query.(string)
}

func (f *indie) recover() {
	f.w.Recover(nil)
}
