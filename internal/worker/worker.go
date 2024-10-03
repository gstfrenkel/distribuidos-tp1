package worker

import (
	"fmt"
	"os"
	"reflect"

	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
)

type Worker interface {
	Init() error
	Start()
}

func Consume(proc func(delivery amqpconn.Delivery), signalChan chan os.Signal, deliveryChan ...<-chan amqpconn.Delivery) {
	cases := make([]reflect.SelectCase, 0, len(deliveryChan)+1)
	cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(signalChan)})

	for _, ch := range deliveryChan {
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
	}

	for {
		chosen, recv, ok := reflect.Select(cases)
		if !ok || chosen == 0 {
			return
		}
		proc(recv.Interface().(amqpconn.Delivery))
	}
}

func InitQueues(b broker.MessageBroker, dsts ...broker.Destination) ([]broker.Queue, []broker.EofDestination, error) {
	var queues []broker.Queue
	var destinations []broker.EofDestination

	for _, dst := range dsts {
		queue, destination, err := initQueue(b, dst)
		if err != nil {
			return nil, nil, err
		}
		queues = append(queues, queue...)
		destinations = append(destinations, destination...)
	}

	return queues, destinations, nil
}

func initQueue(b broker.MessageBroker, dst broker.Destination) ([]broker.Queue, []broker.EofDestination, error) {
	if dst.Consumers == 0 {
		q, err := b.QueueDeclare(dst.Name)
		if err != nil {
			return nil, nil, err
		}
		if err = b.QueueBind(broker.QueueBind{Exchange: dst.Exchange, Name: dst.Name, Key: dst.Key}); err != nil {
			return nil, nil, err
		}
		return q, []broker.EofDestination{{Exchange: dst.Exchange, Key: dst.Key}}, nil
	}

	queues := make([]broker.Queue, 0, dst.Consumers)
	destinations := make([]broker.EofDestination, 0, dst.Consumers)
	for i := uint8(0); i < dst.Consumers; i++ {
		name := fmt.Sprintf(dst.Name, i+1)
		q, err := b.QueueDeclare(name)
		if err != nil {
			return nil, nil, err
		}
		key := fmt.Sprintf(dst.Key, i)
		if err = b.QueueBind(broker.QueueBind{Exchange: dst.Exchange, Name: name, Key: key}); err != nil {
			return nil, nil, err
		}
		queues = append(queues, q...)
		destinations = append(destinations, broker.EofDestination{Exchange: dst.Exchange, Key: key})
	}

	return queues, destinations, nil
}
