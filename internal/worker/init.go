package worker

import (
	"fmt"
	"strings"

	"tp1/pkg/amqp"
)

func (f *Worker) initExchanges() error {
	var exchanges []amqp.Exchange
	if err := f.config.Unmarshal(exchangesKey, &exchanges); err != nil {
		return err
	}

	return f.Broker.ExchangeDeclare(exchanges...)
}

func (f *Worker) initQueues() error {
	if err := f.initOutputQueues(); err != nil {
		return err
	}

	return f.initInputQueues()
}

func (f *Worker) initOutputQueues() error {
	// Output queue unmarshalling.
	if err := f.config.Unmarshal(outputQKey, &f.Outputs); err != nil {
		return err
	}

	for _, output := range f.Outputs {
		// Queue declaration and binding.
		_, destinations, err := f.initQueue(output)
		if err != nil {
			return err
		}
		// EOF Output queue processing.
		for _, dst := range destinations {
			f.outputsEof = append(f.outputsEof, amqp.DestinationEof(dst))
		}
	}

	return nil
}

func (f *Worker) initInputQueues() error {
	// Input queue unmarshalling and binding.
	var inputQ []amqp.Destination
	err := f.config.Unmarshal(inputQKey, &inputQ)
	if err != nil {
		return err
	}

	for _, q := range inputQ {
		if q.Exchange == "" {
			continue
		}

		f.inputEof = amqp.DestinationEof(q)

		if isInputDestinationScalable(q) {
			q.Name = fmt.Sprintf(q.Name, f.Id)
			q.Key = fmt.Sprintf(q.Key, f.Id)
		}

		if _, err = f.Broker.QueueDeclare(q.Name); err != nil {
			return err
		}
		if err = f.Broker.QueueBind(amqp.QueueBind{Exchange: q.Exchange, Name: q.Name, Key: q.Key}); err != nil {
			return err
		}
	}

	return nil
}

func (f *Worker) initQueue(dst amqp.Destination) ([]amqp.Queue, []amqp.Destination, error) {
	if dst.Consumers == 0 {
		return f.initNonScalableQueue(dst)
	}

	return f.initScalableQueue(dst)
}

func (f *Worker) initNonScalableQueue(dst amqp.Destination) ([]amqp.Queue, []amqp.Destination, error) {
	q, err := f.Broker.QueueDeclare(dst.Name)
	if err != nil {
		return nil, nil, err
	}
	if err = f.Broker.QueueBind(amqp.QueueBind{Exchange: dst.Exchange, Name: dst.Name, Key: dst.Key}); err != nil {
		return nil, nil, err
	}
	return q, []amqp.Destination{{Exchange: dst.Exchange, Key: dst.Key}}, nil
}

func (f *Worker) initScalableQueue(dst amqp.Destination) ([]amqp.Queue, []amqp.Destination, error) {
	queues := make([]amqp.Queue, 0, dst.Consumers)
	destinations := make([]amqp.Destination, 0, dst.Consumers)

	for i := uint8(0); i < dst.Consumers; i++ {
		name := fmt.Sprintf(dst.Name, i)
		q, err := f.Broker.QueueDeclare(name)
		if err != nil {
			return nil, nil, err
		}

		key := fmt.Sprintf(dst.Key, i)
		if err = f.Broker.QueueBind(amqp.QueueBind{Exchange: dst.Exchange, Name: name, Key: key}); err != nil {
			return nil, nil, err
		}

		queues = append(queues, q...)

		if isEofOutputDestination(dst, i) {
			destinations = append(destinations, amqp.Destination{Exchange: dst.Exchange, Key: key})
		}
	}

	return queues, destinations, nil
}

func isInputDestinationScalable(dst amqp.Destination) bool {
	return strings.Contains(dst.Name, manyConsumersSubstr) && strings.Contains(dst.Key, manyConsumersSubstr)
}

func isEofOutputDestination(dst amqp.Destination, id uint8) bool {
	return !dst.Single || id == 0
}
