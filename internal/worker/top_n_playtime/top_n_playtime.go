package top_n_playtime

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type filter struct {
	w *worker.Worker
	n int
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

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, nil, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}
	} else if messageId == message.GameIdMsg {
		msg, err := message.DateFilteredReleasesFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(msg message.DateFilteredReleases) {

	// TODO read from json as filter attribute
	n:= 10
	topNPlaytime := msg.ToTopNPlaytimeMessage(n) 
	b, err := topNPlaytime.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	headers := map[string]any{amqp.MessageIdHeader: message.PlatformID}
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
