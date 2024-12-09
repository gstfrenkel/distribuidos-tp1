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
	expectedEofsKey     = "expected-eofs"
	inputQKey           = "input-queues"
	manyConsumersSubstr = "%d"
	exchangesKey        = "exchanges"
	outputQKey          = "output-queues"
)

type Node interface {
	Init() error
	Start()
	Process(delivery amqp.Delivery, headers amqp.Header) ([]sequence.Destination, []byte)
}

// Worker represents a worker node which processes messages from a broker, manages its state,
// and communicates with other system components.
type Worker struct {
	config        config.Config
	Broker        amqp.MessageBroker
	inputEof      amqp.DestinationEof
	outputsEof    []amqp.DestinationEof
	Outputs       []amqp.Destination
	recovery      *recovery.Handler
	sequenceIdGen *sequence.Generator
	dup           *dup.Handler
	Uuid          string
	Id            uint8
	peers         uint8
	ExpectedEofs  uint8
	Query         any
	signalChan    chan os.Signal
}

// New initializes and returns a new instance of Worker.
// It loads the configuration, sets up dependencies like the message broker,
// recovery handler, and health check service, and configures signal handling
// for graceful shutdowns.
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

	return &Worker{
		config:        cfg,
		Query:         query,
		Broker:        b,
		signalChan:    signalChan,
		Uuid:          os.Getenv(workerUuidKey),
		Id:            uint8(id),
		recovery:      recoveryHandler,
		dup:           dup.NewHandler(),
		sequenceIdGen: sequence.NewGenerator(),
		peers:         peers,
		ExpectedEofs:  expectedEofs,
	}, nil
}

// Init initializes the Worker instance by setting up exchanges and queues.
func (f *Worker) Init() error {
	if err := f.initExchanges(); err != nil {
		return err
	}
	return f.initQueues()
}

// Start begins the main processing logic for the Worker.
// It listens for messages from input queues, applies the provided filter logic, and processes messages.
func (f *Worker) Start(filter Node) {
	defer close(f.signalChan)
	defer f.Broker.Close()

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
			f.sequenceIdGen.RecoverId(seq)
		}

		f.dup.RecoverSequenceId(*src)
	}
}

// NextSequenceId generates and returns the next sequence ID for a given key.
func (f *Worker) NextSequenceId(key string) uint64 {
	return f.sequenceIdGen.NextId(key)
}

// HandleEofMessage processes an EOF message and determines the next action based on the workers visited.
//
// This method processes EOF messages, tracks which workers have already handled the EOF, and decides whether
// to forward the message to the next input queue or propagate it to the output queues. If all workers have been
// visited, the EOF message is finalized and sent to the outputs.
//
// Parameters:
// - msg ([]byte): The raw message containing EOF information.
// - headers (amqp.Header): The headers associated with the message.
// - output (...amqp.DestinationEof): A variadic parameter of EOF destinations for output queues.
func (f *Worker) HandleEofMessage(msg []byte, headers amqp.Header, output ...amqp.DestinationEof) ([]sequence.Destination, error) {
	workersVisited, err := message.EofFromBytes(msg)
	if err != nil {
		return nil, err
	}

	if !workersVisited.Contains(f.Id) {
		workersVisited = append(workersVisited, f.Id)
	}

	if uint8(len(workersVisited)) < f.peers {
		return f.sendEofToNextInput(headers, workersVisited)
	}

	return f.sendEofToOutputs(headers, output...)
}

// sendEofToNextInput forwards an EOF message to the next input queue after processing.
//
// This method is responsible for forwarding the EOF message to the next input worker in the sequence.
// It generates a new sequence ID for the next worker and publishes the updated EOF message with the
// appropriate headers, including the worker's sequence ID and message ID.
//
// Parameters:
// - headers (amqp.Header): The headers of the incoming EOF message that will be included in the forwarded message.
// - workersVisited (message.Eof): A list of worker IDs that have already processed the EOF message.
func (f *Worker) sendEofToNextInput(headers amqp.Header, workersVisited message.Eof) ([]sequence.Destination, error) {
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
		headers.WithMessageId(message.EofId).WithSequenceId(sequence.SrcNew(f.Uuid, sequenceId)),
	)
}

// sendEofToOutputs forwards an EOF message to the output queues after all workers have been visited.
//
// This method is responsible for publishing the EOF message to the designated output queues once all workers
// have processed the message. It generates a new sequence ID for each output queue and sends an empty EOF message
// to those queues with the appropriate headers, including the worker's sequence ID and message ID.
//
// Parameters:
// - headers (amqp.Header): The headers of the incoming EOF message that will be included in the forwarded message.
// - output (...amqp.DestinationEof): A variadic parameter of EOF destinations for output queues.
func (f *Worker) sendEofToOutputs(headers amqp.Header, output ...amqp.DestinationEof) ([]sequence.Destination, error) {
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
			headers.WithMessageId(message.EofId).WithSequenceId(sequence.SrcNew(f.Uuid, sequenceId)),
		); err != nil {
			return nil, err
		}
	}

	return sequenceIds, nil
}

// consume listens for incoming AMQP messages and processes them using the provided filter and sequence handling logic.
// The function utilizes a `select` loop to wait for either signal interrupts or incoming messages. Once a message is
// received, it is processed and acknowledged. Duplicate messages are filtered using a handler for sequence IDs.
//
// Parameters:
// - filter (Node): A filter function that processes the incoming message and determines how it should be handled.
// - signalChan (chan os.Signal): A channel that listens for shutdown signals (e.g., SIGINT, SIGTERM).
// - deliveryChan (...<-chan amqp.Delivery): One or more AMQP channels that deliver incoming messages to be processed.
//
// This method performs the following tasks:
// 1. It sets up a `select` loop that listens on the signal channel and the provided delivery channels.
// 2. Upon receiving a signal, the worker shuts down gracefully.
// 3. If an incoming message is received, it checks for duplicates by inspecting the sequence ID.
// 4. Non-duplicate messages are processed by the `filter` and logged using the recovery handler.
// 5. After processing, the message is acknowledged, and the worker continues to wait for the next message or signal.
//
// It ensures that the worker can shut down cleanly when receiving a signal and that messages are processed in order,
// with duplicate handling based on sequence IDs.
func (f *Worker) consume(filter Node, signalChan chan os.Signal, deliveryChan ...<-chan amqp.Delivery) {
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

		// Node and only process non-duplicate messages
		if !f.dup.IsDuplicate(*srcSequenceId) {
			sequenceIds, msg := filter.Process(delivery, header)

			if err = f.recovery.Log(recovery.NewRecord(header, sequenceIds, msg)); err != nil {
				logs.Logger.Errorf("%s: %s", errors.FailedToLog.Error(), err)
			}
		}

		// Acknowledge all duplicate and processed messages
		if err = delivery.Ack(false); err != nil {
			logs.Logger.Errorf("Failed to acknowledge message: %s", err.Error())
		}
	}
}
