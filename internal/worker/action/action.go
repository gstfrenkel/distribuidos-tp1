package action

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

func (f *filter) Process(delivery amqp.Delivery, header amqp.Header) {
	var sequenceIds []sequence.Destination

	switch header.MessageId {
	case message.EofMsg:
		headersEof[amqp.ClientIdHeader] = header.ClientId
		//headersEof[amqp.OriginIdHeader] = f.w.NextSequenceId()

		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, headersEof, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}

		//f.w.Recovery.Log(recovery.NewRecord(header))
	case message.GameIdMsg:
		msg, err := message.GameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		headers[amqp.ClientIdHeader] = delivery.Headers[amqp.ClientIdHeader]
		sequenceIds = f.publish(msg)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), header.MessageId)
	}

	if err := f.w.Recovery.Log(recovery.NewRecord(header, sequenceIds, []string{})); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToLog.Error(), err)
	}
}

func (f *filter) publish(msg message.Game) []sequence.Destination {
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
	go f.w.Recovery.Recover(ch)

	for _, l := range <-ch {

	}

	return nil
}
*/
