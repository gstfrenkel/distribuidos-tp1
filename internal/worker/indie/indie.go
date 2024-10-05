package indie

import (
	"fmt"
	"os"
	"strconv"
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/message"
)

const (
	playtimeQueue = iota
	positiveQueue
)

var (
	routes = [2]broker.Destination{}
	genre  = "indie"
)

type filter struct {
	config     config.Config
	broker     broker.MessageBroker
	input      broker.Route
	outputs    []broker.Route
	signalChan chan os.Signal
	id         uint8
	peers      uint8
}

func New() (worker.Worker, error) {
	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	b, err := amqpconn.NewBroker()
	if err != nil {
		return nil, err
	}

	id, _ := strconv.Atoi(os.Getenv("worker-id"))

	return &filter{
		id:         uint8(id),
		peers:      uint8(cfg.Int("exchange.peers", 1)),
		config:     cfg,
		broker:     b,
		signalChan: worker.SignalChannel(),
	}, nil
}

func (f *filter) Init() error {
	exchange := f.config.String("exchange.name", "indies")

	if err := f.broker.ExchangeDeclare(map[string]string{exchange: f.config.String("exchange.kind", "direct")}); err != nil {
		return err
	}

	if err := f.broker.QueueBind(broker.QueueBind{
		Exchange: exchange,
		Name:     f.config.String("gateway.queue", "games_indie"),
		Key:      f.config.String("gateway.key", "input"),
	}); err != nil {
		return err
	}

	_, outputs, err := worker.InitQueues(f.broker, []broker.Destination{{
		Exchange: exchange,
		Key:      f.config.String("playtime-queue.key", ""),
		Name:     f.config.String("playtime-queue.name", "indie_games"),
	}, {
		Exchange:  exchange,
		Key:       f.config.String("positive-queue.key", "%d"),
		Name:      f.config.String("positive-queue.name", "query3_games_%d"),
		Consumers: f.config.Uint8("positive-queue.consumers", 0),
	}}...)

	if err != nil {
		return err
	}

	f.outputs = outputs
	f.input = broker.Route{Exchange: exchange, Key: f.config.String("gateway.key", "input")}
	routes[playtimeQueue] = broker.Destination{
		Exchange: exchange,
		Key:      f.config.String("playtime-queue.key", ""),
	}
	routes[positiveQueue] = broker.Destination{
		Exchange:  exchange,
		Consumers: f.config.Uint8("positive-queue.consumers", 0),
		Key:       f.config.String("positive-queue.key", "%d"),
	}
	return nil
}

func (f *filter) Start() {
	defer f.broker.Close()

	ch, err := f.broker.Consume(f.config.String("gateway.queue", "games_indie"), "", true, false)
	if err != nil {
		println(err.Error())
		return
	}

	worker.Consume(f.process, f.signalChan, ch)
}

func (f *filter) process(reviewDelivery amqpconn.Delivery) {
	messageId := message.ID(reviewDelivery.Headers[amqpconn.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.broker.HandleEofMessage(f.id, f.peers, reviewDelivery.Body, f.input, f.outputs...); err != nil {
			fmt.Printf("\n%s\n", errors.FailedToPublish.Error())
		}
	} else if messageId == message.GameIdMsg {
		msg, err := message.GameFromBytes(reviewDelivery.Body)
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg)
	} else {
		fmt.Printf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish(msg message.Game) {
	gameReleases := msg.ToGameReleasesMessage(genre)
	b, err := gameReleases.ToBytes()
	if err != nil {

	}

	if err = f.broker.Publish(routes[playtimeQueue].Exchange, routes[playtimeQueue].Key, uint8(message.GameReleaseID), b); err != nil {
		fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	gameNames := msg.ToGameNamesMessage(genre)
	for _, game := range gameNames {
		b, err = game.ToBytes()
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := routes[positiveQueue].Key
		if routes[positiveQueue].Consumers > 0 {
			k = worker.ShardGameId(game.GameId, k, routes[positiveQueue].Consumers)
		}
		if err = f.broker.Publish(routes[positiveQueue].Exchange, k, uint8(message.GameNameID), b); err != nil {
			fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}

}
