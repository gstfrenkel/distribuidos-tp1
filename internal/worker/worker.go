package worker

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"

	"tp1/pkg/broker"
	"tp1/pkg/broker/amqpconn"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"

	"github.com/pierrec/xxHash/xxHash32"
)

type Filter interface {
	Init() error
	Start()
	Process(delivery amqpconn.Delivery)
}

type Worker struct {
	config     config.Config
	Query      any
	Broker     broker.MessageBroker
	InputEof   broker.DestinationEof
	OutputsEof []broker.DestinationEof
	Outputs    []broker.Destination
	SignalChan chan os.Signal
	Id         uint8
	Peers      uint8
}

func New() (*Worker, error) {
	cfg, err := provider.LoadConfig("config.json")
	if err != nil {
		return nil, err
	}
	_ = logs.InitLogger(cfg.String("log-level", "INFO"))
	b, err := amqpconn.NewBroker()
	if err != nil {
		return nil, err
	}

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	id, _ := strconv.Atoi(os.Getenv("worker-id"))
	var query any
	if err = cfg.Unmarshal("query", &query); err != nil {
		return nil, err
	}

	return &Worker{
		config:     cfg,
		Query:      query,
		Broker:     b,
		SignalChan: signalChan,
		Id:         uint8(id),
	}, nil
}

func (f *Worker) Init() error {
	if err := f.initExchanges(); err != nil {
		return err
	}
	return f.initQueues()
}

func (f *Worker) Start(filter Filter) {
	defer f.Broker.Close()

	var inputQ broker.Destination
	err := f.config.Unmarshal("inputQ", &inputQ)
	if err != nil {
		logs.Logger.Errorf("error unmarshalling input-queue: %s", err.Error())
		return
	}

	ch, err := f.Broker.Consume(inputQ.Name, "", true, false)
	if err != nil {
		logs.Logger.Errorf("error consuming from input-queue: %s", err.Error())
		return
	}

	f.consume(filter, f.SignalChan, ch)
}

func (f *Worker) consume(filter Filter, signalChan chan os.Signal, deliveryChan ...<-chan amqpconn.Delivery) {
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
		filter.Process(recv.Interface().(amqpconn.Delivery))
	}
}

func (f *Worker) initExchanges() error {
	var exchanges []broker.Exchange
	if err := f.config.Unmarshal("exchanges", &exchanges); err != nil {
		return err
	}

	return f.Broker.ExchangeDeclare(exchanges...)
}

func (f *Worker) initQueues() error {
	// Input queue unmarshaling and binding.
	if err := f.config.Unmarshal("input-queues", &f.InputEof); err != nil {
		return err
	} else if err = f.Broker.QueueBind(broker.QueueBind{Exchange: f.InputEof.Exchange, Name: f.InputEof.Name, Key: f.InputEof.Key}); err != nil {
		return err
	}

	// Output queue unmarshaling.
	if err := f.config.Unmarshal("output-queues", &f.Outputs); err != nil {
		return err
	}

	var outputsEof []broker.DestinationEof

	for _, dst := range f.Outputs {
		// Queue declaration and binding.
		_, destination, err := f.initQueue(dst)
		if err != nil {
			return err
		}
		// EOF Output queue processing.
		for _, aux := range destination {
			outputsEof = append(outputsEof, broker.DestinationEof(aux))
		}
	}

	return nil
}

func (f *Worker) initQueue(dst broker.Destination) ([]broker.Queue, []broker.Destination, error) {
	if dst.Consumers == 0 {
		q, err := f.Broker.QueueDeclare(dst.Name)
		if err != nil {
			return nil, nil, err
		}
		if err = f.Broker.QueueBind(broker.QueueBind{Exchange: dst.Exchange, Name: dst.Name, Key: dst.Key}); err != nil {
			return nil, nil, err
		}
		return q, []broker.Destination{{Exchange: dst.Exchange, Key: dst.Key}}, nil
	}

	queues := make([]broker.Queue, 0, dst.Consumers)
	destinations := make([]broker.Destination, 0, dst.Consumers)

	for i := uint8(0); i < dst.Consumers; i++ {
		name := fmt.Sprintf(dst.Name, i)
		q, err := f.Broker.QueueDeclare(name)
		if err != nil {
			return nil, nil, err
		}
		key := fmt.Sprintf(dst.Key, i)
		if err = f.Broker.QueueBind(broker.QueueBind{Exchange: dst.Exchange, Name: name, Key: key}); err != nil {
			return nil, nil, err
		}
		queues = append(queues, q...)
		destinations = append(destinations, broker.Destination{Exchange: dst.Exchange, Key: key})
	}

	return queues, destinations, nil
}

func ShardGameId(id int64, key string, consumers uint8) string {
	if consumers == 0 {
		return key
	}
	return fmt.Sprintf(key, xxHash32.Checksum([]byte{byte(id)}, 0)%uint32(consumers))
}
