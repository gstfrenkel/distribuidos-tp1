package client

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type Client struct {
	cfg          config.Config
	sigChan      chan os.Signal
	stopped      bool
	stoppedMutex sync.Mutex
	resultsFile  *os.File
	clientId     string
}

func New() (*Client, error) {
	config, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	return &Client{
		cfg:          config,
		sigChan:      sigChan,
		stopped:      false,
		stoppedMutex: sync.Mutex{},
	}, nil
}

func (c *Client) Start() {
	logs.Logger.Info("Client running...")

	wg := sync.WaitGroup{}
	gatewayAddress := c.cfg.String("gateway.address", "172.25.125.100")

	err := c.fetchClientID(gatewayAddress)
	if err != nil {
		logs.Logger.Errorf("Error fetching client ID: %v", err)
		return
	}

	err = c.openResultsFile()
	if err != nil {
		logs.Logger.Errorf("Error opening results file: %v", err)
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.startResultsListener(gatewayAddress)
	}()

	gamesPort := c.cfg.String("gateway.games_port", "5051")
	gamesFullAddress := gatewayAddress + ":" + gamesPort

	reviewsPort := c.cfg.String("gateway.reviews_port", "5050")
	reviewsFullAddress := gatewayAddress + ":" + reviewsPort

	gamesConn, err := net.Dial("tcp", gamesFullAddress)
	if err != nil {
		logs.Logger.Errorf("Games Conn error:", err)
		return
	}

	reviewsConn, err := net.Dial("tcp", reviewsFullAddress)
	if err != nil {
		logs.Logger.Errorf("Reviews Conn error:", err)
		return
	}

	logs.Logger.Infof("Games conn: %s", gamesFullAddress)
	logs.Logger.Infof("Reviews conn: %s", reviewsFullAddress)

	wg.Add(2)

	go func() {
		<-c.sigChan
		c.handleSigterm()
	}()

	go func() {
		defer wg.Done()
		defer gamesConn.Close()
		readAndSendCSV(c.cfg.String("client.games_path", "data/games.csv"), uint8(message.GameIdMsg), gamesConn, &message.DataCSVGames{}, c)
		header := make([]byte, 32)
		if _, err = gamesConn.Read(header); err != nil {
			logs.Logger.Errorf("Failed to read message: %v", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		defer reviewsConn.Close()
		readAndSendCSV(c.cfg.String("client.reviews_path", "data/reviews.csv"), uint8(message.ReviewIdMsg), reviewsConn, &message.DataCSVReviews{}, c)
		header := make([]byte, 32)
		if _, err = reviewsConn.Read(header); err != nil {
			logs.Logger.Errorf("Failed to read message: %v", err.Error())
		}
	}()

	wg.Wait()
	logs.Logger.Info("Client exit...")
	c.Close(gamesConn, reviewsConn)
}
