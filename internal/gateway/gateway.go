package gateway

import (
	"encoding/binary"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tp1/pkg/ioutils"
	"tp1/pkg/message"

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
	Config      config.Config
	broker      amqp.MessageBroker
	queues      []amqp.Queue //order: reviews, games_platform, games_action, games_indie
	exchange    string
	Listeners   [connections]net.Listener
	ChunkChans  [connections]chan ChunkItem
	finished    bool
	finishedMu  sync.Mutex
	resultsChan chan []byte
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
	GatewayExchangeName, err := rabbit.CreateExchange(cfg, b, cfg.String("rabbitmq.exchange_name", "gateway"))
	if err != nil {
		return nil, err
	}

	err = rabbit.BindGatewayQueuesToExchange(b, queues, cfg, GatewayExchangeName)
	if err != nil {
		return nil, err
	}

	// Reports exchange
	ReportsExchangeName, err := rabbit.CreateExchange(cfg, b, cfg.String("rabbitmq.reports", "reports"))
	if err != nil {
		return nil, err
	}

	err = rabbit.BindReportsQueueToExchange(b, cfg, ReportsExchangeName)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		Config:      cfg,
		broker:      b,
		queues:      queues,
		exchange:    GatewayExchangeName,
		ChunkChans:  [connections]chan ChunkItem{make(chan ChunkItem), make(chan ChunkItem)},
		finished:    false,
		finishedMu:  sync.Mutex{},
		Listeners:   [connections]net.Listener{},
		resultsChan: make(chan []byte),
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

	go startChunkSender(GamesListener, g.ChunkChans[GamesListener], g.broker, g.exchange, g.Config.Uint8("gateway.chunk_size", 100), g.Config.String("rabbitmq.games_routing_key", "game"))
	go startChunkSender(ReviewsListener, g.ChunkChans[ReviewsListener], g.broker, g.exchange, g.Config.Uint8("gateway.chunk_size", 100), g.Config.String("rabbitmq.reviews_routing_key", "review"))
	go ListenResults(g)

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

func (g *Gateway) HandleResults() error {

	//logs.Logger.Info("Conectting to client...")
	clientAddr := g.Config.String("client_addr", "172.25.125.99:5050")
	cliConn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		logs.Logger.Errorf("Error connecting to client address %s: %s", clientAddr, err)
		return err
	}
	defer cliConn.Close()

	for {
		select {
		case rabbitMsg := <-g.resultsChan:
			clientMsg := message.ClientMessage{
				DataLen: uint32(len(rabbitMsg)),
				Data:    rabbitMsg,
			}

			data := make([]byte, 4+len(clientMsg.Data))
			binary.BigEndian.PutUint32(data[:4], clientMsg.DataLen)
			copy(data[4:], clientMsg.Data)

			if err := ioutils.SendAll(cliConn, data); err != nil {
				logs.Logger.Errorf("Error sending message to client: %s", err)
				return err
			}

			//logs.Logger.Infof("Sent message to client: %s", string(rabbitMsg))
		}
	}
}
