package indie

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

const (
	query2 uint8 = iota
	query3
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
	if err := f.w.Init(); err != nil {
		return err
	}

	f.recover()

	return nil
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
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

func (f *filter) publish(headers amqp.Header, msg message.Game) []sequence.Destination {
	genre := getGenre(f)
	sequenceIdsRelease := f.publishGameReleases(headers, msg, genre)
	sequenceIdsNames := f.publishGameNames(headers, msg, genre)
	return append(sequenceIdsRelease, sequenceIdsNames...)
}

func (f *filter) publishGameNames(headers amqp.Header, msg message.Game, genre string) []sequence.Destination {
	gameNames := msg.ToGameNamesMessage(genre)
	sequenceIdsNames := make([]sequence.Destination, 0, len(gameNames)*len(f.w.Outputs))
	headers = headers.WithMessageId(message.GameNameID)

	for _, game := range gameNames {
		b, err := game.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := worker.ShardGameId(game.GameId, f.w.Outputs[query3].Key, f.w.Outputs[query3].Consumers)
		sequenceId := f.w.NextSequenceId(k)
		sequenceIdsNames = append(sequenceIdsNames, sequence.DstNew(k, sequenceId))

		if err = f.w.Broker.Publish(f.w.Outputs[query3].Exchange, k, b, headers.WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}

	return sequenceIdsNames
}

func (f *filter) publishGameReleases(headers amqp.Header, msg message.Game, genre string) []sequence.Destination {
	gameReleases := msg.ToGameReleasesMessage(genre)
	sequenceId := f.w.NextSequenceId(f.w.Outputs[query2].Key)
	sequenceIds := []sequence.Destination{sequence.DstNew(f.w.Outputs[query2].Key, sequenceId)}
	headers = headers.WithMessageId(message.GameReleaseID).WithSequenceId(sequence.SrcNew(f.w.Id, sequenceId))

	b, err := gameReleases.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return sequenceIds
	}

	if err = f.w.Broker.Publish(f.w.Outputs[query2].Exchange, f.w.Outputs[query2].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	return sequenceIds
}

func getGenre(f *filter) string {
	return f.w.Query.(string)
}

func (f *filter) recover() {
	f.w.Recover(nil)
}
