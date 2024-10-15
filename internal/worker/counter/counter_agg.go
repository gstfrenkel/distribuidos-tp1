package counter

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type filter struct {
	w         *worker.Worker
	games     message.GameNames
	batchSize uint16
	eofsRecv  uint8
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{w: w, games: nil}, nil
}

func (f *filter) Init() error {
	f.batchSize = uint16(f.w.Query.(float64))
	return f.w.Init()
}

func (f *filter) Start() {
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		f.eofsRecv++
		if f.eofsRecv >= f.w.Peers {
			f.publish()
		}
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}
		f.games = append(f.games, msg)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish() {
	if f.games == nil {
		return
	}

	worker.SendBatches(f.games.ToAny(), f.batchSize, message.GameNameFromAnyToBytes, f.sendBatch)
	logs.Logger.Infof("Q4 games: %v", f.games)
	f.sendEof()
	f.reset()
}

func (f *filter) sendBatch(b []byte) {
	headers := map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID), amqp.OriginIdHeader: amqp.Query4originId}
	if err := f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}

func (f *filter) sendEof() {
	headers := map[string]any{amqp.OriginIdHeader: amqp.Query4originId}
	if err := f.w.Broker.HandleEofMessage(f.w.Id, 0, amqp.EmptyEof, headers, f.w.InputEof, f.w.OutputsEof...); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Eof message sent")
}

func (f *filter) reset() {
	f.games = nil
	f.eofsRecv = 0
}
