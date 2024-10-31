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
	games     map[string]message.GameNames // <clientId, []GameName>
	batchSize uint16
	eofsRecv  map[string]uint8 // <clientId, eofsReceived>
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}

	return &filter{
		w:        w,
		games:    make(map[string]message.GameNames),
		eofsRecv: make(map[string]uint8),
	}, nil
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
	clientId := delivery.Headers[amqp.ClientIdHeader].(string)

	if messageId == message.EofMsg {
		f.eofsRecv[clientId]++
		if f.eofsRecv[clientId] >= f.w.Peers {
			f.publish(clientId, true)
			f.sendEof(clientId)
			f.reset(clientId)
		}
	} else if messageId == message.GameNameID {
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.saveGame(msg, clientId)
		f.publish(clientId, false)
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(clientId string, eof bool) {
	if len(f.games[clientId]) > 0 && (len(f.games[clientId]) >= int(f.batchSize) || eof) {
		b, err := f.games[clientId].ToBytes()
		logs.Logger.Infof("Publishing batch of %d games for client %s", len(f.games), clientId)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.sendBatch(b, clientId)
		f.games[clientId] = f.games[clientId][:0]
	}
}

func (f *filter) sendBatch(b []byte, clientId string) {
	headers := map[string]any{amqp.MessageIdHeader: uint8(message.GameNameID), amqp.OriginIdHeader: amqp.Query4originId, amqp.ClientIdHeader: clientId}
	if err := f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}

func (f *filter) sendEof(clientId string) {
	headers := map[string]any{amqp.OriginIdHeader: amqp.Query4originId, amqp.ClientIdHeader: clientId}
	if err := f.w.Broker.HandleEofMessage(f.w.Id, 0, amqp.EmptyEof, headers, f.w.InputEof, f.w.OutputsEof...); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Eof message sent for client %s", clientId)
}

func (f *filter) reset(clientId string) {
	delete(f.eofsRecv, clientId)
	delete(f.games, clientId)
}

func (f *filter) saveGame(msg message.GameName, clientId string) {
	if _, ok := f.games[clientId]; !ok {
		f.games[clientId] = make(message.GameNames, 0)
	}

	f.games[clientId] = append(f.games[clientId], msg)
}
