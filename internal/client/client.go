package client

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const (
	configPath        = "config.toml"
	gatewayAddrKey    = "gateway.address"
	gatewayAddrDef    = "172.25.125.100"
	signals           = 2
	gamesPortKey      = "gateway.games_port"
	gamesPortDef      = "5051"
	reviewsPortKey    = "gateway.reviews_port"
	reviewsPortDef    = "5050"
	gamesCsvPathKey   = "client.games_path"
	gamesCsvPathDef   = "data/games.csv"
	reviewsCsvPathKey = "client.reviews_path"
	reviewsCsvPathDef = "data/reviews.csv"
	csvsToSend        = 2
	ackBytes          = 32
)

type Client struct {
	cfg          config.Config
	sigChan      chan os.Signal
	stopped      bool
	stoppedMutex sync.Mutex
	resultsFile  *os.File
	clientId     string
	gatewayAddr  string
}

func New() (*Client, error) {
	config, err := provider.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	sigChan := make(chan os.Signal, signals)
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
	go func() {
		<-c.sigChan
		c.handleSigterm()
	}()

	wg := sync.WaitGroup{}
	gatewayAddress := c.cfg.String(gatewayAddrKey, gatewayAddrDef)

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

	gamesFullAddress := c.getFullAddress(gatewayAddress, gamesPortKey, gamesPortDef)
	reviewsFullAddress := c.getFullAddress(gatewayAddress, reviewsPortKey, reviewsPortDef)

	gamesConn, err := c.setupConnection(gamesFullAddress)
	if err != nil {
		return
	}

	reviewsConn, err := c.setupConnection(reviewsFullAddress)
	if err != nil {
		return
	}

	wg.Add(csvsToSend)

	c.sendGames(&wg, gamesConn, gamesFullAddress)
	c.sendReviews(&wg, reviewsConn, reviewsFullAddress)

	wg.Wait()
	logs.Logger.Info("Client exit...")
	c.Close(gamesConn, reviewsConn)
}

func (c *Client) sendData(wg *sync.WaitGroup, conn net.Conn, pathKey string, pathDef string, id uint8, dataStruct interface{}, address string) {
	go func() {
		defer wg.Done()
		defer conn.Close()
		c.readAndSendCSV(c.cfg.String(pathKey, pathDef), id, conn, dataStruct, address)
	}()
}

func (c *Client) sendGames(wg *sync.WaitGroup, gamesConn net.Conn, address string) {
	c.sendData(wg, gamesConn, gamesCsvPathKey, gamesCsvPathDef, uint8(message.GameIdMsg), &message.DataCSVGames{}, address)
}

func (c *Client) sendReviews(wg *sync.WaitGroup, reviewsConn net.Conn, address string) {
	c.sendData(wg, reviewsConn, reviewsCsvPathKey, reviewsCsvPathDef, uint8(message.ReviewIdMsg), &message.DataCSVReviews{}, address)
}

func (c *Client) getFullAddress(gatewayAddress, portKey, portDef string) string {
	port := c.cfg.String(portKey, portDef)
	return gatewayAddress + ":" + port
}

func (c *Client) reconnect(address string, timeout int) net.Conn {

	var newConn net.Conn
	for {
		logs.Logger.Infof("Attempting to reconnect...")
		conn, err := c.setupConnection(address)
		if err == nil {
			logs.Logger.Infof("Reconnected successfully.")
			newConn = conn
			break
		}
		logs.Logger.Errorf("Reconnect failed, retrying in %v seconds: %s", timeout, err)
		time.Sleep(time.Duration(timeout) * time.Second)
	}
	return newConn
}

func (c *Client) setupConnection(address string) (net.Conn, error) {
	logs.Logger.Infof("Connecting to %s...", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		logs.Logger.Info("Failed to connect to %s: %v", address, err)
		return nil, err
	}

	logs.Logger.Infof("Connection established: %s", address)

	err = c.sendClientID(conn)
	if err != nil {
		logs.Logger.Errorf("Error sending client ID: %s", err)
		conn.Close()
		return nil, err
	}

	return conn, nil
}
