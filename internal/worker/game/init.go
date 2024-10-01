package game

import (
	"fmt"
	"tp1/pkg/broker"
)

func (f filter) queues() []string {
	baseName := f.config.String("queue.name", "shooters_%d")
	consumers := f.config.Int("queue.consumers", 1)
	queues := make([]string, 0, consumers)

	for i := 1; i <= consumers; i++ {
		queues = append(queues, fmt.Sprintf(baseName, i))
	}

	return queues
}

func (f filter) binds() ([]broker.QueueBind, broker.Destination, []broker.Destination) {
	ex := f.config.String("exchange", "shooter")
	baseName := f.config.String("queue.name", "shooters_%d")
	baseKey := f.config.String("queue.key", "%d")
	consumers := f.config.Int("queue.consumers", 1)

	input := broker.Destination{
		Exchange: ex,
		Key:      f.config.String("gateway.key", "input"),
	}
	outputs := make([]broker.Destination, 0, consumers)
	binds := make([]broker.QueueBind, 0, consumers+1)
	binds = append(binds, broker.QueueBind{
		Name:     f.config.String("gateway.queue", "games_shooter"),
		Key:      input.Key,
		Exchange: input.Exchange,
	})

	for i := 1; i <= consumers; i++ {
		key := fmt.Sprintf(baseKey, i-1)
		binds = append(binds, broker.QueueBind{Name: fmt.Sprintf(baseName, i), Key: key, Exchange: ex})
		outputs = append(outputs, broker.Destination{Exchange: ex, Key: key})
	}

	return binds, input, outputs
}
