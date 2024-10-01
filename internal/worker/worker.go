package worker

import (
	"os"
	"reflect"
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
