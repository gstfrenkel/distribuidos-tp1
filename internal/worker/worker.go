package worker

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"tp1/internal/errors"
	"tp1/pkg/amqp"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/dup"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"

	"github.com/pierrec/xxHash/xxHash32"
)

type Filter interface {
	Init() error
	Start()
	Process(delivery amqp.Delivery, header amqp.Header) ([]sequence.Destination, []string)
}

type Worker struct {
	config      config.Config
	Query       any
	Broker      amqp.MessageBroker
	inputEof    amqp.DestinationEof
	outputsEof  []amqp.DestinationEof
	Outputs     []amqp.Destination
	signalChan  chan os.Signal
	recovery    recovery.Handler
	dup         dup.Handler
	sequenceIds map[string]uint64
	Id          uint8
	Peers       uint8
}

func New() (*Worker, error) {
	cfg, err := provider.LoadConfig("config.json")
	if err != nil {
		return nil, err
	}
	_ = logs.InitLogger(cfg.String("log-level", "INFO"))
	b, err := broker.NewBroker()
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

	var peers uint8
	if err = cfg.Unmarshal("peers", &peers); err != nil {
		return nil, err
	}

	recoveryHandler, err := recovery.NewHandler()
	if err != nil {
		return nil, err
	}

	return &Worker{
		config:      cfg,
		Query:       query,
		Broker:      b,
		signalChan:  signalChan,
		Id:          uint8(id),
		recovery:    recoveryHandler,
		dup:         dup.NewHandler(),
		sequenceIds: make(map[string]uint64),
		Peers:       peers,
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
	defer close(f.signalChan)

	var inputQ []amqp.Destination
	err := f.config.Unmarshal("input-queues", &inputQ)
	if err != nil {
		logs.Logger.Errorf("error unmarshalling input-queue: %s", err.Error())
		return
	}

	channels := make([]<-chan amqp.Delivery, 0, len(inputQ))
	for _, q := range inputQ {
		queueName := q.Name
		if strings.Contains(queueName, "%d") {
			queueName = fmt.Sprintf(queueName, f.Id)
		}
		if _, err = f.Broker.QueueDeclare(queueName); err != nil {
			logs.Logger.Errorf("error declaring queue %s: %s", queueName, err.Error())
		}
		ch, err := f.Broker.Consume(queueName, "", false, false)
		if err != nil {
			logs.Logger.Errorf("error consuming from input-queue: %s", err.Error())
			return
		}
		channels = append(channels, ch)
	}

	f.consume(filter, f.signalChan, channels...)
}

func (f *Worker) NextSequenceId(key string) uint64 {
	sequenceId, ok := f.sequenceIds[key]
	if !ok {
		f.sequenceIds[key] = 1
	} else {
		f.sequenceIds[key]++
	}
	return sequenceId
}

func (f *Worker) HandleEofMessage(msg []byte, headers map[string]any, output ...amqp.DestinationEof) ([]sequence.Destination, error) {
	workersVisited, err := message.EofFromBytes(msg)
	if err != nil {
		return nil, err
	}

	if !workersVisited.Contains(f.Id) {
		workersVisited = append(workersVisited, f.Id)
	}

	if headers == nil {
		headers = map[string]any{amqp.MessageIdHeader: uint8(message.EofMsg)}
	} else {
		headers[amqp.MessageIdHeader] = uint8(message.EofMsg)
	}

	var sequenceIds []sequence.Destination

	if uint8(len(workersVisited)) < f.Peers {
		sequenceIds = append(sequenceIds, sequence.DstNew(f.inputEof.Key, f.NextSequenceId(f.inputEof.Key)))

		bytes, err := workersVisited.ToBytes()
		if err != nil {
			return nil, err
		}
		return sequenceIds, f.Broker.Publish(f.inputEof.Exchange, f.inputEof.Key, bytes, headers)
	}

	outputs := f.outputsEof
	if output != nil && len(output) > 0 {
		outputs = output
	}

	for _, o := range outputs {
		sequenceIds = append(sequenceIds, sequence.DstNew(o.Key, f.NextSequenceId(o.Key)))
		if err = f.Broker.Publish(o.Exchange, o.Key, amqp.EmptyEof, headers); err != nil {
			return nil, err
		}
	}

	return sequenceIds, nil
}

func (f *Worker) consume(filter Filter, signalChan chan os.Signal, deliveryChan ...<-chan amqp.Delivery) {
	cases := make([]reflect.SelectCase, 0, len(deliveryChan)+1)
	cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(signalChan)})

	for _, ch := range deliveryChan {
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
	}

	for {
		chosen, recv, ok := reflect.Select(cases)
		if !ok || chosen == 0 { // Signal channel chosen for consumption
			logs.Logger.Criticalf("Signal received. Shutting down...")
			return
		}

		delivery := recv.Interface().(amqp.Delivery)
		header := amqp.HeadersFromDelivery(delivery)
		srcSequenceId, err := sequence.SrcFromString(header.SequenceId)
		if err != nil {
			logs.Logger.Errorf("error getting source sequence id: %s", err.Error())
			continue
		}

		// Filter and only process non-duplicate messages
		if !f.dup.IsDuplicate(*srcSequenceId) {
			sequenceIds, msg := filter.Process(delivery, header)

			if err = f.recovery.Log(recovery.NewRecord(header, sequenceIds, msg)); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToLog.Error(), err)
			}
		} else {
			logs.Logger.Infof("Received duplicate: %v", srcSequenceId)
		}

		// Acknowledge all duplicate and processed messages
		if err = delivery.Ack(false); err != nil {
			logs.Logger.Errorf("Failed to acknowledge message: %s", err.Error())
		}
	}
}

func (f *Worker) initExchanges() error {
	var exchanges []amqp.Exchange
	if err := f.config.Unmarshal("exchanges", &exchanges); err != nil {
		return err
	}

	return f.Broker.ExchangeDeclare(exchanges...)
}

func (f *Worker) initQueues() error {
	// Output queue unmarshalling.
	if err := f.config.Unmarshal("output-queues", &f.Outputs); err != nil {
		return err
	}

	for _, dst := range f.Outputs {
		// Queue declaration and binding.
		_, destination, err := f.initQueue(dst)
		if err != nil {
			return err
		}
		// EOF Output queue processing.
		for _, aux := range destination {
			f.outputsEof = append(f.outputsEof, amqp.DestinationEof(aux))
		}
	}

	// Input queue unmarshalling and binding.
	var inputQ []amqp.Destination
	err := f.config.Unmarshal("input-queues", &inputQ)
	if err != nil {
		return err
	}

	for _, q := range inputQ {
		if q.Exchange != "" {
			if _, err = f.Broker.QueueDeclare(q.Name); err != nil {
				return err
			}
			f.inputEof = amqp.DestinationEof(q)
			if err = f.Broker.QueueBind(amqp.QueueBind{Exchange: f.inputEof.Exchange, Name: f.inputEof.Name, Key: f.inputEof.Key}); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (f *Worker) initQueue(dst amqp.Destination) ([]amqp.Queue, []amqp.Destination, error) {
	if dst.Consumers == 0 {
		q, err := f.Broker.QueueDeclare(dst.Name)
		if err != nil {
			return nil, nil, err
		}
		if err = f.Broker.QueueBind(amqp.QueueBind{Exchange: dst.Exchange, Name: dst.Name, Key: dst.Key}); err != nil {
			return nil, nil, err
		}
		return q, []amqp.Destination{{Exchange: dst.Exchange, Key: dst.Key}}, nil
	}

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
		destinations = append(destinations, amqp.Destination{Exchange: dst.Exchange, Key: key})
	}

	return queues, destinations, nil
}

func ShardGameId(id int64, key string, consumers uint8) string {
	if consumers == 0 {
		return key
	}
	return fmt.Sprintf(key, xxHash32.Checksum([]byte{byte(id)}, 0)%uint32(consumers))
}
