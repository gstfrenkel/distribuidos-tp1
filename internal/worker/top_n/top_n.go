package top_n

import (
	"tp1/internal/worker"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/message"
)

type filter struct {
	w *worker.Worker
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

func (f *filter) Process(reviewDelivery broker.Delivery) {
	messageId := message.ID(reviewDelivery.Headers[broker.MessageIdHeader].(uint8))
	// TODO
}

func (f *filter) publish(msg message.Game) {
	//TODO
}
