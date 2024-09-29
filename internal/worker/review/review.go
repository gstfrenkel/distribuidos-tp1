package review

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
	outputExchange = "reviews"

	positiveConsumers = 1
	negativeConsumers = 1

	positiveKey = "p%d"
	negativeKey = "n%d"

	input   broker.Aaa
	outputs []broker.Aaa
)

type Filter struct {
	config     config.Config
	broker     broker.MessageBroker
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

	return &Filter{
		id:         uint8(id),
		peers:      uint8(cfg.Int("exchange.peers", 1)),
		config:     cfg,
		broker:     b,
		signalChan: signalChan,
	}, nil
}

func (f Filter) Init() error {
	outputExchange = f.config.String("exchange.name", "reviews")
	positiveKey = f.config.String("positive-reviews-sh.key", "p%d")
	negativeKey = f.config.String("negative-reviews-sh.key", "n%d")
	positiveConsumers = f.config.Int("positive-reviews-sh.consumers", 1)
	negativeConsumers = f.config.Int("negative-reviews-sh.consumers", 1)

	if err := f.broker.ExchangeDeclare(broker.Exchange{Name: outputExchange, Kind: f.config.String("outputExchange.kind", "direct")}); err != nil {
		return err
	} else if _, err = f.broker.QueueDeclare(f.queues()...); err != nil {
		return err
	} else if err = f.broker.QueueBind(f.binds()...); err != nil {
		return err
	}

	input = broker.Aaa{Exchange: f.config.String("gateway.exchange", "reviews"), Key: f.config.String("gateway.key", "reviews")}
	outputs = append(outputs, broker.Aaa{Exchange: outputExchange, Key: ""})
	for i := 0; i < positiveConsumers; i++ {
		outputs = append(outputs, broker.Aaa{Exchange: outputExchange, Key: fmt.Sprintf(positiveKey, i)})
	}
	for i := 0; i < negativeConsumers; i++ {
		outputs = append(outputs, broker.Aaa{Exchange: outputExchange, Key: fmt.Sprintf(negativeKey, i)})
	}

	return nil
}

func (f Filter) Start() {
	defer f.broker.Close()

	reviewChan, err := f.broker.Consume(f.config.String("gateway.queue-name", "reviews"), "", true, false)
	if err != nil {
		println(err.Error())
		return
	}

	for {
		select {
		case reviewDelivery, ok := <-reviewChan:
			if !ok {
				return
			}
			f.process(reviewDelivery)
		case sig := <-f.signalChan:
			fmt.Printf("Received signal: %s. Shutting down...", sig.String())
			return
		}
	}
}

func (f Filter) process(reviewDelivery amqpconn.Delivery) {
	messageId := message.ID(reviewDelivery.Headers[amqpconn.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {
		if err := f.broker.HandleEofMessage(f.id, f.peers, reviewDelivery.Body, input, outputs...); err != nil {
			fmt.Printf("\n%s\n", errors.FailedToPublish.Error())
		}
	} else if messageId == message.ReviewIdMsg {
		msg, err := message.ReviewFromBytes(reviewDelivery.Body)
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}

		f.publish(msg)
	} else {
		fmt.Printf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f Filter) publish(msg message.Review) {
	b, err := msg.ToPositiveReviewWithTextMessage().ToBytes()
	if err != nil {
		fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
	} else if err = f.broker.Publish(outputExchange, "", uint8(message.ReviewID), b); err != nil {
		fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
	}

	f.shardPublish(msg.ToPositiveReviewMessage(), positiveKey, positiveConsumers)
	f.shardPublish(msg.ToNegativeReviewMessage(), negativeKey, negativeConsumers)
}

func (f Filter) shardPublish(reviews message.ScoredReviews, k string, consumers int) {
	for _, rv := range reviews {
		b, err := rv.ToBytes()
		if err != nil {
			fmt.Printf("%s: %s", errors.FailedToParse.Error(), err.Error())
			continue
		}

		key := fmt.Sprintf(k, rv.GameId%int64(consumers))
		if err = f.broker.Publish(outputExchange, key, uint8(message.ReviewID), b); err != nil {
			fmt.Printf("%s: %s", errors.FailedToPublish.Error(), err.Error())
		}
	}
}
