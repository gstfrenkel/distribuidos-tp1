package gateway

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"tp1/internal/gateway/id_generator"
	"tp1/internal/healthcheck"
	"tp1/pkg/message"

	"tp1/internal/gateway/rabbit"
	"tp1/pkg/amqp"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
)

const (
	configFilePath        = "config.toml"
	GamesListener         = 0
	ReviewsListener       = 1
	ResultsListener       = 2
	ClientIdListener      = 3
	connections           = 4
	chunkChans            = 2
	exchangeNameKey       = "rabbitmq.exchange_name"
	exchangeDefault       = "gateway"
	reportsKey            = "rabbitmq.reports"
	reportsDefault        = "reports"
	workerIdKey           = "worker-id"
	chunkSizeKey          = "gateway.chunk_size"
	chunkSizeDefault      = 100
	gamesRoutingKey       = "rabbitmq.games_routing_key"
	gamesRoutingDefault   = "game"
	reviewsRoutingKey     = "rabbitmq.reviews_routing_key"
	reviewsRoutingDefault = "review"
	signals               = 2
)

type Gateway struct {
	Config                   config.Config
	broker                   amqp.MessageBroker
	queues                   []amqp.Queue //order: reviews, games_platform, games_action, games_indie
	exchange                 string
	reportsExchange          string
	Listeners                [connections]net.Listener
	ChunkChans               [chunkChans]chan ChunkItem
	finished                 bool
	finishedMu               sync.Mutex
	IdGenerator              *id_generator.IdGenerator
	IdGeneratorMu            sync.Mutex
	clientChannels           sync.Map
	clientGamesAckChannels   sync.Map
	clientReviewsAckChannels sync.Map
	healthCheckService       *healthcheck.Service
}

func New() (*Gateway, error) {
	cfg, err := provider.LoadConfig(configFilePath)
	if err != nil {
		return nil, err
	}

	b, err := broker.NewBroker()
	if err != nil {
		return nil, err
	}

	queues, err := rabbit.CreateGatewayQueues(b, cfg)
	if err != nil {
		return nil, err
	}

	// Gateway exchange
	GatewayExchangeName, err := rabbit.CreateExchange(cfg, b, cfg.String(exchangeNameKey, exchangeDefault))
	if err != nil {
		return nil, err
	}

	// Reports exchange
	ReportsExchangeName, err := rabbit.CreateExchange(cfg, b, cfg.String(reportsKey, reportsDefault))
	if err != nil {
		return nil, err
	}

	err = rabbit.BindGatewayQueuesToExchange(b, queues, cfg, GatewayExchangeName, ReportsExchangeName)
	if err != nil {
		return nil, err
	}

	gId, err := strconv.Atoi(os.Getenv(workerIdKey))
	if err != nil {
		return nil, err
	}

	hc, err := healthcheck.NewService()
	if err != nil {
		return nil, err
	}

	return &Gateway{
		Config:                   cfg,
		broker:                   b,
		queues:                   queues,
		exchange:                 GatewayExchangeName,
		reportsExchange:          ReportsExchangeName,
		ChunkChans:               [chunkChans]chan ChunkItem{make(chan ChunkItem), make(chan ChunkItem)},
		finished:                 false,
		finishedMu:               sync.Mutex{},
		Listeners:                [connections]net.Listener{},
		IdGenerator:              id_generator.New(uint8(gId)),
		IdGeneratorMu:            sync.Mutex{},
		clientChannels:           sync.Map{},
		clientGamesAckChannels:   sync.Map{},
		clientReviewsAckChannels: sync.Map{},
		healthCheckService:       hc,
	}, nil
}

func (g *Gateway) Start() {
	defer g.broker.Close()
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
		g.ChunkChans[GamesListener], g.broker, g.exchange,
		g.Config.Uint8(chunkSizeKey, chunkSizeDefault),
		g.Config.String(gamesRoutingKey, gamesRoutingDefault),
	)

	go startChunkSender(ReviewsListener, &g.clientReviewsAckChannels,
		g.ChunkChans[ReviewsListener], g.broker, g.exchange,
		g.Config.Uint8(chunkSizeKey, chunkSizeDefault),
		g.Config.String(reviewsRoutingKey, reviewsRoutingDefault),
	)

	go g.ListenResults()

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

func matchMessageId(listener int) message.ID {
	if listener == ReviewsListener {
		return message.ReviewIdMsg
	}
	return message.GameIdMsg
}

func matchListenerId(msgId message.ID) int {
	if msgId == message.ReviewIdMsg {
		return ReviewsListener
	}
	return GamesListener
}

func (g *Gateway) free(sigs chan os.Signal) {
	g.broker.Close()
	g.Listeners[ReviewsListener].Close()
	g.Listeners[GamesListener].Close()
	g.Listeners[ResultsListener].Close()
	g.Listeners[ClientIdListener].Close()
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
