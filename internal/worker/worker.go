package worker

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"

	"github.com/pierrec/xxHash/xxHash32"
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

func InitQueues(b broker.MessageBroker, dsts ...broker.Destination) ([]broker.Queue, []broker.Route, error) {
	var queues []broker.Queue
	var destinations []broker.Route

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

func initQueue(b broker.MessageBroker, dst broker.Destination) ([]broker.Queue, []broker.Route, error) {
	if dst.Consumers == 0 {
		q, err := b.QueueDeclare(dst.Name)
		if err != nil {
			return nil, nil, err
		}
		if err = b.QueueBind(broker.QueueBind{Exchange: dst.Exchange, Name: dst.Name, Key: dst.Key}); err != nil {
			return nil, nil, err
		}
		return q, []broker.Route{{Exchange: dst.Exchange, Key: dst.Key}}, nil
	}

	queues := make([]broker.Queue, 0, dst.Consumers)
	destinations := make([]broker.Route, 0, dst.Consumers)
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
		destinations = append(destinations, broker.Route{Exchange: dst.Exchange, Key: key})
	}

	return queues, destinations, nil
}

func SignalChannel() chan os.Signal {
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	return signalChan
}

func ShardGameId(id int64, key string, consumers uint8) string {
	return fmt.Sprintf(key, xxHash32.Checksum([]byte{byte(id)}, 0)%uint32(consumers))
}
