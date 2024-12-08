package aggregator

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/sequence"
)

type counter struct {
	agg   *aggregator
	games map[string]message.GameNames // <clientId, []GameName>
}

func NewCounter() (worker.Node, error) {
	a, err := newAggregator(amqp.Query4OriginId)
	if err != nil {
		return nil, err
	}
	a.batchSize = uint16(a.w.Query.(float64))

	return &counter{
		agg:   a,
		games: make(map[string]message.GameNames),
	}, nil
}

func (c *counter) Init() error {
	if err := c.agg.w.Init(); err != nil {
		return err
	}

	c.agg.recover(c, message.GameNameId)

	return nil
}

func (c *counter) Start() {
	c.agg.w.Start(c)
}

func (c *counter) Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte, bool) {
	var sequenceIds []sequence.Destination
	var done bool

	switch headers.MessageId {
	case message.EofId:
		sequenceIds, done = c.agg.processEof(c, headers.WithOriginId(c.agg.originId), false)
	case message.GameNameId:
		c.save(delivery.Body, headers.ClientId)
		sequenceIds = c.publish(headers)
	default:
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), headers.MessageId)
	}

	return sequenceIds, delivery.Body, done
}

func (c *counter) publish(headers amqp.Header) []sequence.Destination {
	var sequenceIds []sequence.Destination
	if len(c.games[headers.ClientId]) > 0 && (c.batchSizeReached(headers) || c.agg.eofsReached(headers)) {
		b, err := c.games[headers.ClientId].ToBytes()
		logs.Logger.Infof("Publishing batch of %d games for client %s", len(c.games), headers.ClientId)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return sequenceIds
		}

		sequenceIds = c.sendBatch(headers, b)
		c.games[headers.ClientId] = c.games[headers.ClientId][:0]
	}

	return sequenceIds
}

func (c *counter) sendBatch(headers amqp.Header, b []byte) []sequence.Destination {
	output := shardOutput(c.agg.w.Outputs[0], headers.ClientId)
	key := output.Key
	sequenceId := c.agg.w.NextSequenceId(key, headers.ClientId)
	headers = headers.WithSequenceId(sequence.SrcNew(c.agg.w.Uuid, sequenceId)).WithOriginId(c.agg.originId)

	if err := c.agg.w.Broker.Publish(output.Exchange, key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}

	return []sequence.Destination{sequence.DstNew(key, sequenceId)}
}

func (c *counter) reset(clientId string) {
	delete(c.games, clientId)
}

func (c *counter) save(msgBytes []byte, clientId string) {
	msg, err := message.GameNameFromBytes(msgBytes)
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
		return
	}

	if _, ok := c.games[clientId]; !ok {
		c.games[clientId] = make(message.GameNames, 0)
	}

	c.games[clientId] = append(c.games[clientId], msg)
}

func (c *counter) batchSizeReached(headers amqp.Header) bool {
	return len(c.games[headers.ClientId]) >= int(c.agg.batchSize)
}
