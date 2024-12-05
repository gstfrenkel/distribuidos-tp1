package gateway

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"tp1/internal/gateway/rabbit"
	"tp1/internal/healthcheck"
	"tp1/pkg/amqp"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/dup"
	"tp1/pkg/logs"
	"tp1/pkg/message"
	"tp1/pkg/recovery"
	"tp1/pkg/utils/id"
)

const (
	configFilePath   = "config.toml"
	GamesListener    = 0
	ReviewsListener  = 1
	ResultsListener  = 2
	ClientIdListener = 3
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
	ChunkChans               [chunkChans]chan ChunkItem
	finished                 bool
	finishedMu               sync.Mutex
	IdGenerator              *id.Generator
	IdGeneratorMu            sync.Mutex
	clientChannels           sync.Map
	clientGamesAckChannels   sync.Map
	clientReviewsAckChannels sync.Map
	healthCheckService       healthcheck.Service
	recovery                 recovery.Handler
	logChannel               chan recovery.Record
	dup                      dup.Handler
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
		ChunkChans:               [chunkChans]chan ChunkItem{make(chan ChunkItem), make(chan ChunkItem)},
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

	go startChunkSender(GamesListener, &g.clientGamesAckChannels,
		g.ChunkChans[GamesListener], g.broker, g.destinations[1:],
		g.Config.Uint8(chunkSizeKey, chunkSizeDefault),
	)

	go startChunkSender(ReviewsListener, &g.clientReviewsAckChannels,
		g.ChunkChans[ReviewsListener], g.broker, g.destinations[0:1],
		g.Config.Uint8(chunkSizeKey, chunkSizeDefault),
	)

	go g.ListenResults()

	go g.logResults()

	wg := &sync.WaitGroup{}
	wg.Add(connections)
	g.startListeners(wg, err)
	wg.Wait()

	g.free(sigs)
}

func (g *Gateway) startListeners(wg *sync.WaitGroup, err error) {
	go func() {
		defer wg.Done()
		err = g.listenForNewClient()
		if err != nil {
			logs.Logger.Errorf("Error listening reviews: %s", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		err = g.listenForData(ReviewsListener)
		if err != nil {
			logs.Logger.Errorf("Error listening reviews: %s", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		err = g.listenForData(GamesListener)
		if err != nil {
			logs.Logger.Errorf("Error listening games: %s", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		err = g.listenResultsRequests()
		if err != nil {
			logs.Logger.Errorf("Error listening results: %s", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		g.healthCheckService.Listen()
	}()
}

func matchMessageId(listener int) message.Id {
	if listener == ReviewsListener {
		return message.ReviewId
	}
	return message.GameId
}

func matchListenerId(msgId message.Id) int {
	if msgId == message.ReviewId {
		return ReviewsListener
	}
	return GamesListener
}

func (g *Gateway) free(sigs chan os.Signal) {
	g.broker.Close()
	_ = g.Listeners[ReviewsListener].Close()
	_ = g.Listeners[GamesListener].Close()
	_ = g.Listeners[ResultsListener].Close()
	_ = g.Listeners[ClientIdListener].Close()
	g.healthCheckService.Close()
	close(g.ChunkChans[ReviewsListener])
	close(g.ChunkChans[GamesListener])
	close(sigs)
}

func (g *Gateway) HandleSIGTERM() {
	logs.Logger.Info("Received SIGTERM, shutting down")
	g.finishedMu.Lock()
	g.finished = true
	g.finishedMu.Unlock()
}
