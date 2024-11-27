package counter

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
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

func (f *filter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte) {
	var sequenceIds []sequence.Destination

	headers = headers.WithOriginId(amqp.Query4originId)

	switch headers.MessageId {
	case message.EofMsg:
		f.eofsRecv[headers.ClientId]++
		if f.eofsRecv[headers.ClientId] >= f.w.ExpectedEofs {
			f.publish(headers, true)
			f.sendEof(headers)
			f.reset(headers.ClientId)
		}
	case message.GameNameID:
		msg, err := message.GameNameFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		} else {
			f.saveGame(msg, headers.ClientId)
			f.publish(headers, false)
		}
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, nil
}

func (f *filter) publish(headers amqp.Header, eof bool) {
	if len(f.games[headers.ClientId]) > 0 && (len(f.games[headers.ClientId]) >= int(f.batchSize) || eof) {
		b, err := f.games[headers.ClientId].ToBytes()
		logs.Logger.Infof("Publishing batch of %d games for client %s", len(f.games), headers.ClientId)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.sendBatch(headers, b)
		f.games[headers.ClientId] = f.games[headers.ClientId][:0]
	}
}

func (f *filter) sendBatch(headers amqp.Header, b []byte) {
	headers = headers.WithMessageId(message.GameNameID)
	if err := f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}

func (f *filter) sendEof(headers amqp.Header) {
	_, err := f.w.HandleEofMessage(amqp.EmptyEof, headers)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	logs.Logger.Infof("Eof message sent for client %s", headers.ClientId)
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
