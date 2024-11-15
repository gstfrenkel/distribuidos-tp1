package release_date

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
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
	return f.w.Init()
}

func (f *filter) Start() {
	slice := f.w.Query.([]any)
	f.startYear = int(slice[0].(float64))
	f.endYear = int(slice[1].(float64))
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery, _ amqp.Header) {
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, map[string]any{amqp.ClientIdHeader: clientId}, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	} else if messageId == message.GameReleaseID {
		msg, err := message.ReleasesFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg, clientId)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(msg message.Releases, clientId string) {
	dateFilteredGames := msg.ToPlaytimeMessage(f.startYear, f.endYear)

	b, err := dateFilteredGames.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	headers := map[string]any{
		amqp.MessageIdHeader: uint8(message.GameWithPlaytimeID),
		amqp.ClientIdHeader:  clientId,
	}

	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
