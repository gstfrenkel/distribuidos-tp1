package gateway

import (
	"net"

	"tp1/internal/gateway/rabbit"
	"tp1/pkg/amqp"
	"tp1/pkg/amqp/broker"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
)

const configFilePath = "config.toml"

type Gateway struct {
	Config    config.Config
	broker    amqp.MessageBroker
	queues    []amqp.Queue //order: reviews, games_platform, games_shooter, games_indie
	exchange  string
	Listener  net.Listener
	ChunkChan chan ChunkItem
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
		Config:    cfg,
		broker:    b,
		queues:    queues,
		exchange:  exchangeName,
		ChunkChan: make(chan ChunkItem),
	}, nil
}

func (g Gateway) Start() {
	defer g.broker.Close()

	err := CreateGatewaySocket(&g)
	if err != nil {
		logs.Logger.Errorf("Failed to create gateway socket: %s", err.Error())
		return
	}

	defer g.Listener.Close() //TODO handle

	go startChunkSender(g.ChunkChan, g.broker, g.exchange, g.Config.Uint8("gateway.chunk_size", 100), g.Config.String("rabbitmq.reviews_routing_key", "review"), g.Config.String("rabbitmq.games_routing_key", "game"))

	err = ListenForNewClients(&g)
	if err != nil {
		logs.Logger.Errorf("Failed to listen for new clients: %s", err.Error())
		return
	}
}

func (g Gateway) End() {
	g.broker.Close()
	//TODO , cerrar chan
}
