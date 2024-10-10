package gateway

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"tp1/internal/gateway/rabbit"
	"tp1/pkg/amqp"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
)

const configFilePath = "config.toml"
const GamesListener = 0
const ReviewsListener = 1
const connections = 2

type Gateway struct {
	Config     config.Config
	broker     amqp.MessageBroker
	queues     []amqp.Queue //order: reviews, games_platform, games_action, games_indie
	exchange   string
	Listeners  [2]net.Listener
	ChunkChan  chan ChunkItem
	finished   bool
	finishedMu sync.Mutex
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

	exchangeName, err := rabbit.CreateGatewayExchange(cfg, b)
	if err != nil {
		return nil, err
	}

	err = rabbit.BindGatewayQueuesToExchange(b, queues, cfg, exchangeName)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		Config:     cfg,
		broker:     b,
		queues:     queues,
		exchange:   exchangeName,
		ChunkChan:  make(chan ChunkItem),
		finished:   false,
		finishedMu: sync.Mutex{},
		Listeners:  [connections]net.Listener{},
	}, nil
}

func (g *Gateway) Start() {
	defer g.broker.Close()
	sigs := make(chan os.Signal, 2)
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

	go startChunkSender(g.ChunkChan, g.broker, g.exchange, g.Config.Uint8("gateway.chunk_size", 100), g.Config.String("rabbitmq.reviews_routing_key", "review"), g.Config.String("rabbitmq.games_routing_key", "game"))

	wg := &sync.WaitGroup{}
	wg.Add(connections)

	go func() {
		defer wg.Done()
		err = g.listenForNewClients(ReviewsListener)
		if err != nil {
			logs.Logger.Errorf("Error listening reviews: %s", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		err = g.listenForNewClients(GamesListener)
		if err != nil {
			logs.Logger.Errorf("Error listening games: %s", err.Error())
		}
	}()

	wg.Wait()
	g.free(sigs)
}

func (g *Gateway) free(sigs chan os.Signal) {
	g.broker.Close()
	g.Listeners[ReviewsListener].Close()
	g.Listeners[GamesListener].Close()
	close(g.ChunkChan)
	close(sigs)
}

func (g *Gateway) HandleSIGTERM() {
	g.finishedMu.Lock()
	g.finished = true
	g.finishedMu.Unlock()
}
