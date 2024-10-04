package platform

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/message"
)

var (
	route = broker.Destination{}
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

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	id, _ := strconv.Atoi(os.Getenv("worker-id"))

	return &filter{
		id:         uint8(id),
		peers:      uint8(cfg.Int("exchange.peers", 1)),
		config:     cfg,
		broker:     b,
		signalChan: signalChan,
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

	_, output, err := worker.InitQueues(f.broker, []broker.Destination{{
		Exchange: exchange,
		Key:      f.config.String("count-queue.key", ""),
		Name:     f.config.String("count-queue.name", "indie_games"),
	}}...)

	if err != nil {
		return err
	}

	f.outputs = output
	f.input = broker.Route{Exchange: exchange, Key: f.config.String("gateway.key", "input")}
	route = broker.Destination{
		Exchange: exchange,
		Key:      f.config.String("count-queue.key", ""),
	}
	return nil
}

func (f *filter) Start() {
	defer f.broker.Close()

	ch, err := f.broker.Consume(f.config.String("gateway.queue", "games_shooter"), "", true, false)
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
	platforms := msg.ToPlatformMessage()
	b, err := platforms.ToBytes()
	if err != nil {
		fmt.Printf("\n%s\n", errors.FailedToParse.Error())
	}

	if err := f.broker.Publish(route.Exchange, route.Key, uint8(message.PlatformID), b); err != nil {
		fmt.Printf("\n%s\n", errors.FailedToPublish.Error())
	}
}
