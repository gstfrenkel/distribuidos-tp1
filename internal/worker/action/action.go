package action

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

var (
	headersEof = map[string]any{amqp.OriginIdHeader: amqp.GameOriginId}
	headers    = map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID)}
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

	return nil
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
	headers[amqp.ClientIdHeader] = header.ClientId

	games := msg.ToGameNamesMessage(f.w.Query.(string))
	sequenceIds := make([]sequence.Destination, 0, len(games)*len(f.w.Outputs))

	for _, game := range games {
		b, err := game.ToBytes()
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		for _, output := range f.w.Outputs {
			key := worker.ShardGameId(game.GameId, output.Key, output.Consumers)
			sequenceId := f.w.NextSequenceId(key)
			sequenceIds = append(sequenceIds, sequence.DstNew(key, sequenceId))
			headers[amqp.SequenceIdHeader] = sequence.SrcNew(f.w.Id, sequenceId).ToString()

			if err = f.w.Broker.Publish(output.Exchange, key, b, headers); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
			}
		}
	}

	return sequenceIds
}

/*func (f *filter) recover() {
	ch := make(chan []string)
	go f.w.recoveryt.Recover(ch)

	for _, l := range <-ch {

	}

	return nil
}
*/
