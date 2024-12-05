package gateway

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"tp1/internal/gateway/chunk"
	"tp1/internal/gateway/persistence"
	"tp1/internal/gateway/rabbit"
	"tp1/internal/gateway/utils"
	"tp1/internal/healthcheck"
	"tp1/pkg/amqp"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/dup"
	"tp1/pkg/logs"
	"tp1/pkg/recovery"
	"tp1/pkg/utils/id"
)

const (
	configFilePath   = "config.toml"
	connections      = 4
	chunkChans       = 2
	exchangeNameKey  = "rabbitmq.exchange_name"
	workerIdKey      = "worker-id"
	chunkSizeKey     = "gateway.chunk_size"
	chunkSizeDefault = 100
	signals          = 2
)

type Gateway struct {
	Config                   config.Config
	broker                   amqp.MessageBroker
	queues                   []amqp.Queue //order: reviews, games_platform, games_action, games_indie
	destinations             []amqp.Destination
	exchange                 string
	Listeners                [connections]net.Listener
	ChunkChans               [chunkChans]chan chunk.Item
	finished                 bool
	finishedMu               sync.Mutex
	IdGenerator              *id.Generator
	IdGeneratorMu            sync.Mutex
	clientChannels           sync.Map
	clientGamesAckChannels   sync.Map
	clientReviewsAckChannels sync.Map
	healthCheckService       *healthcheck.Service
	recovery                 *recovery.Handler
	logChannel               chan recovery.Record
	dup                      *dup.Handler
}

func New() (*Gateway, error) {
	cfg, err := provider.LoadConfig(configFilePath)
	if err != nil {
		return nil, err
	}

	gId, err := strconv.Atoi(os.Getenv(workerIdKey))
	if err != nil {
		return nil, err
	}

	b, err := broker.NewBroker()
	if err != nil {
		return nil, err
	}

	destinations, queues, err := rabbit.CreateGatewayQueues(uint8(gId), b, cfg)
	if err != nil {
		return nil, err
	}

	hc, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}

	recoveryHandler, err := recovery.NewHandler()
	if err != nil {
		return nil, err
	}

	return &Gateway{
		Config:                   cfg,
		broker:                   b,
		queues:                   queues,
		exchange:                 cfg.String(exchangeNameKey, ""),
		destinations:             destinations,
		ChunkChans:               [chunkChans]chan chunk.Item{make(chan chunk.Item), make(chan chunk.Item)},
		finished:                 false,
		finishedMu:               sync.Mutex{},
		Listeners:                [connections]net.Listener{},
		IdGenerator:              id.NewGenerator(uint8(gId), ""),
		IdGeneratorMu:            sync.Mutex{},
		clientChannels:           sync.Map{},
		clientGamesAckChannels:   sync.Map{},
		clientReviewsAckChannels: sync.Map{},
		healthCheckService:       hc,
		recovery:                 recoveryHandler,
		logChannel:               make(chan recovery.Record),
		dup:                      dup.NewHandler(),
	}, nil
}

func (g *Gateway) Start() {
	defer g.broker.Close()
	defer g.IdGenerator.Close()

	sigs := make(chan os.Signal, signals)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		g.HandleSIGTERM()
	}()

	err := g.createGatewaySockets()
	if err != nil {
		logs.Logger.Errorf("Failed to create gateway socket: %s", err.Error())
		return
	}

	go chunk.Start(utils.GamesListener, &g.clientGamesAckChannels,
		g.ChunkChans[utils.GamesListener], g.broker, g.destinations[1:],
		g.Config.Uint8(chunkSizeKey, chunkSizeDefault),
	)

	go chunk.Start(utils.ReviewsListener, &g.clientReviewsAckChannels,
		g.ChunkChans[utils.ReviewsListener], g.broker, g.destinations[0:1],
		g.Config.Uint8(chunkSizeKey, chunkSizeDefault),
	)

	go g.ListenResults()

	go g.logResults()

	wg := &sync.WaitGroup{}
	wg.Add(connections)
	g.startListeners(wg)
	wg.Wait()

	g.free(sigs)
}

func (g *Gateway) startListeners(wg *sync.WaitGroup) {
	go g.startNewClientListener(wg)
	go g.startDataListener(wg, utils.ReviewsListener, "reviews")
	go g.startDataListener(wg, utils.GamesListener, "games")
	go g.startResultsListener(wg)
	go g.startHealthCheckListener(wg)
}

func (g *Gateway) startNewClientListener(wg *sync.WaitGroup) {
	defer wg.Done()
	err := g.listenForNewClient()
	if err != nil {
		logs.Logger.Errorf("Error listening new client: %s", err)
	}
}

func (g *Gateway) startDataListener(wg *sync.WaitGroup, listener int, listenerType string) {
	defer wg.Done()
	err := g.listenForData(listener)
	if err != nil {
		logs.Logger.Errorf("Error listening %s: %s", listenerType, err)
	}
}

func (g *Gateway) startResultsListener(wg *sync.WaitGroup) {
	defer wg.Done()
	err := g.listenResultsRequests()
	if err != nil {
		logs.Logger.Errorf("Error listening results: %s", err)
	}
}

func (g *Gateway) startHealthCheckListener(wg *sync.WaitGroup) {
	defer wg.Done()
	g.healthCheckService.Listen()
}

func (g *Gateway) recoverResults(
	ch chan recovery.Record,
	clientAccumulatedResults map[string]map[uint8]string,
	recoveredMessages map[string]map[uint8]string,
) {
	go g.recovery.Recover(ch)

	for recoveredMsg := range ch {
		originId := recoveredMsg.Header().OriginId
		sequenceId := recoveredMsg.Header().SequenceId

		if string(recoveredMsg.Message()) != utils.Ack {
			seqSource, err := sequence.SrcFromString(sequenceId)
			if err != nil {
				logs.Logger.Errorf("Failed to parse sequence source: %v", err)
				continue
			}

			g.dup.Add(*seqSource)
		}

		switch originId {
		case amqp.Query1OriginId, amqp.Query2OriginId, amqp.Query3OriginId:
			persistence.HandleSimpleQueryRecovery(recoveredMsg, recoveredMessages)
		case amqp.Query4OriginId, amqp.Query5OriginId:
			persistence.HandleAccumulatingQueryRecovery(
				recoveredMsg,
				clientAccumulatedResults,
				recoveredMessages,
			)
		default:
			logs.Logger.Infof("Header x-origin-id does not match any known origin IDs, got: %v", originId)
		}
	}
}

func (g *Gateway) logResults() {
	for record := range g.logChannel {
		if err := g.recovery.Log(record); err != nil {
			logs.Logger.Errorf("Failed to log record: %s", err)
		}
	}
}

func (g *Gateway) free(sigs chan os.Signal) {
	g.broker.Close()
	_ = g.Listeners[utils.ReviewsListener].Close()
	_ = g.Listeners[utils.GamesListener].Close()
	_ = g.Listeners[utils.ResultsListener].Close()
	_ = g.Listeners[utils.ClientIdListener].Close()
	g.healthCheckService.Close()
	close(g.ChunkChans[utils.ReviewsListener])
	close(g.ChunkChans[utils.GamesListener])
	close(sigs)
}

func (g *Gateway) HandleSIGTERM() {
	logs.Logger.Info("Received SIGTERM, shutting down")
	g.finishedMu.Lock()
	g.finished = true
	g.finishedMu.Unlock()
}
