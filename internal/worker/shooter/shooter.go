package shooter

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"tp1/internal/errors"
	"tp1/pkg/message"

	"tp1/internal/worker"
	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
)

var (
	exchange = "shooters"
	key      = "%d"
	genre    = "action"
)

type filter struct {
	config     config.Config
	broker     broker.MessageBroker
	input      broker.EofDestination
	outputs    []broker.EofDestination
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
	exchange = f.config.String("exchange.name", exchange)

	if err := f.broker.ExchangeDeclare(map[string]string{exchange: f.config.String("exchange.kind", "direct")}); err != nil {
		return err
	}

	if err := f.broker.QueueBind(broker.QueueBind{
		Exchange: exchange,
		Name:     f.config.String("gateway.queue", "games_shooter"),
		Key:      f.config.String("gateway.key", "input"),
	}); err != nil {
		return err
	}

	_, outputs, err := worker.InitQueues(f.broker, []broker.Destination{{
		Exchange:  exchange,
		Key:       f.config.String("count-queue.key", "q4%d"),
		Name:      f.config.String("count-queue.name", "games_query4_%d"),
		Consumers: f.config.Uint8("count-queue.consumers", 0),
	}, {
		Exchange:  exchange,
		Key:       f.config.String("percentile-queue.key", "q5%d"),
		Name:      f.config.String("percentile-queue.name", "games_query5_%d"),
		Consumers: f.config.Uint8("percentile-queue.consumers", 0),
	}}...)

	if err != nil {
		return err
	}

	f.outputs = outputs
	f.input = broker.EofDestination{Exchange: exchange, Key: f.config.String("gateway.key", "input")}
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
	games := msg.ToGameIdMessage(genre)
	for _, game := range games {
		b, err := game.ToBytes()
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		k := fmt.Sprintf(key, int64(game)%int64(len(f.outputs)))
		if err = f.broker.Publish(exchange, k, uint8(message.GameIdID), b); err != nil {
			fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}
