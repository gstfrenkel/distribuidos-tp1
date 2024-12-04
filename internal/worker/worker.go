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
	"tp1/internal/healthcheck"
	"tp1/pkg/amqp"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/dup"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/sequence"
)

const (
	ChanSize            = 32
	signals             = 2
	configPath          = "config.json"
	logLevelKey         = "log-level"
	defaultLogLevel     = "INFO"
	workerIdKey         = "worker-id"
	workerUuidKey       = "worker-uuid"
	queryKey            = "query"
	peersKey            = "peers"
	expectedEofsKey     = "expected_eofs"
	inputQKey           = "input-queues"
	manyConsumersSubstr = "%d"
	exchangesKey        = "exchanges"
	outputQKey          = "output-queues"
)

type W interface {
	Init() error
	Start()
	Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte)
}

type Worker struct {
	config             config.Config
	Query              any
	Broker             amqp.MessageBroker
	inputEof           amqp.DestinationEof
	outputsEof         []amqp.DestinationEof
	Outputs            []amqp.Destination
	signalChan         chan os.Signal
	recovery           recovery.Handler
	dup                dup.Handler
	sequenceIds        map[string]uint64
	Uuid               string
	Id                 uint8
	peers              uint8
	ExpectedEofs       uint8
	HealthCheckService *healthcheck.Service
}

func New() (*Worker, error) {
	cfg, err := provider.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	_ = logs.InitLogger(cfg.String(logLevelKey, defaultLogLevel))
	b, err := broker.NewBroker()
	if err != nil {
		return nil, err
	}

	signalChan := make(chan os.Signal, signals)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	id, _ := strconv.Atoi(os.Getenv(workerIdKey))
	var query any
	if err = cfg.Unmarshal(queryKey, &query); err != nil {
		return nil, err
	}

	var peers uint8
	if err = cfg.Unmarshal(peersKey, &peers); err != nil {
		return nil, err
	}

	var expectedEofs uint8
	if err = cfg.Unmarshal(expectedEofsKey, &expectedEofs); err != nil {
		return nil, err
	}

	recoveryHandler, err := recovery.NewHandler()
	if err != nil {
		return nil, err
	}

	hc, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}

	return &Worker{
		config:             cfg,
		Query:              query,
		Broker:             b,
		signalChan:         signalChan,
		Uuid:               os.Getenv(workerUuidKey),
		Id:                 uint8(id),
		recovery:           recoveryHandler,
		dup:                dup.NewHandler(),
		sequenceIds:        make(map[string]uint64),
		peers:              peers,
		ExpectedEofs:       expectedEofs,
		HealthCheckService: hc,
	}, nil
}

func (f *Worker) Init() error {
	if err := f.initExchanges(); err != nil {
		return err
	}
	return f.initQueues()
}

func (f *Worker) Start(filter W) {
	defer f.Broker.Close()
	defer close(f.signalChan)

	f.listenHc()

	defer f.HealthCheckService.Close()

	var inputQ []amqp.Destination
	err := f.config.Unmarshal(inputQKey, &inputQ)
	if err != nil {
		logs.Logger.Errorf("error unmarshalling input-queue: %s", err.Error())
		return
	}

	channels := make([]<-chan amqp.Delivery, 0, len(inputQ))
	for _, q := range inputQ {
		queueName := q.Name
		if strings.Contains(queueName, manyConsumersSubstr) {
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

// Recover reads the recovery log and processes the messages.
//
// It performs the following steps:
// - Recover the source sequence id.
// - Recover the destination sequence id.
// - Send the message to the filter for processing if needed.
func (f *Worker) Recover(ch chan<- recovery.Message) {
	if ch != nil {
		defer close(ch)
	}

	recoveryCh := make(chan recovery.Record, ChanSize)
	go f.recovery.Recover(recoveryCh)

	for record := range recoveryCh {
		src, err := sequence.SrcFromString(record.Header().SequenceId)
		if err != nil {
			logs.Logger.Errorf("error getting source from sequence: %s", err.Error())
			continue
		}

		if ch != nil {
			ch <- recovery.NewMessage(record)
		}

		for _, seq := range record.SequenceIds() {
			f.recoverDstSequenceId(seq)
		}

		f.recoverSrcSequenceId(*src)
	}
}

func (f *Worker) recoverSrcSequenceId(source sequence.Source) {
	f.dup.Add(source)
}

func (f *Worker) recoverDstSequenceId(destination sequence.Destination) {
	f.sequenceIds[destination.Key()] = destination.Id() + 1
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

func (f *Worker) HandleEofMessage(msg []byte, headers amqp.Header, output ...amqp.DestinationEof) ([]sequence.Destination, error) {
	workersVisited, err := message.EofFromBytes(msg)
	if err != nil {
		return nil, err
	}

	if !workersVisited.Contains(f.Id) {
		workersVisited = append(workersVisited, f.Id)
	}

	if uint8(len(workersVisited)) < f.peers {
		return f.handleEofToInput(headers, workersVisited)
	}

	return f.handleEofToOutputs(headers, output...)
}

func (f *Worker) handleEofToInput(headers amqp.Header, workersVisited message.Eof) ([]sequence.Destination, error) {
	key := fmt.Sprintf(f.inputEof.Key, f.Id+1)
	sequenceId := f.NextSequenceId(key)
	sequenceIds := []sequence.Destination{sequence.DstNew(key, sequenceId)}

	bytes, err := workersVisited.ToBytes()
	if err != nil {
		return nil, err
	}

	return sequenceIds, f.Broker.Publish(
		f.inputEof.Exchange,
		key,
		bytes,
		headers.WithMessageId(message.EofMsg).WithSequenceId(sequence.SrcNew(f.Uuid, sequenceId)),
	)
}

func (f *Worker) handleEofToOutputs(headers amqp.Header, output ...amqp.DestinationEof) ([]sequence.Destination, error) {
	outputs := f.outputsEof
	if output != nil && len(output) > 0 {
		outputs = output
	}

	sequenceIds := make([]sequence.Destination, 0, len(outputs))

	for _, o := range outputs {
		sequenceId := f.NextSequenceId(o.Key)
		sequenceIds = append(sequenceIds, sequence.DstNew(o.Key, sequenceId))
		if err := f.Broker.Publish(
			o.Exchange,
			o.Key,
			amqp.EmptyEof,
			headers.WithMessageId(message.EofMsg).WithSequenceId(sequence.SrcNew(f.Uuid, sequenceId)),
		); err != nil {
			return nil, err
		}
	}

	return sequenceIds, nil
}

func (f *Worker) listenHc() {
	go func() {
		f.HealthCheckService.Listen()
	}()
}

func (f *Worker) consume(filter W, signalChan chan os.Signal, deliveryChan ...<-chan amqp.Delivery) {
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

		// W and only process non-duplicate messages
		if !f.dup.IsDuplicate(*srcSequenceId) {
			sequenceIds, msg := filter.Process(delivery, header)

			if err := f.recovery.Log(recovery.NewRecord(header, sequenceIds, msg)); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToLog.Error(), err)
			}
		}

		// Acknowledge all duplicate and processed messages
		if err := delivery.Ack(false); err != nil {
			logs.Logger.Errorf("Failed to acknowledge message: %s", err.Error())
		}
	}
}

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
