package client

import (
	"fmt"
	"io"
	"log"
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
	done := make(chan bool)

	wg.Add(1)
	go func() {
		c.startListener(&wg, done)
	}()

	address := c.cfg.String("gateway.address", "172.25.125.100")

	gamesPort := c.cfg.String("gateway.games_port", "5051")
	gamesFullAddress := address + ":" + gamesPort

	reviewsPort := c.cfg.String("gateway.reviews_port", "5050")
	reviewsFullAddress := address + ":" + reviewsPort

	gamesConn, err := net.Dial("tcp", gamesFullAddress)
	if err != nil {
		fmt.Println("Games Conn error:", err)
		return
	}

	reviewsConn, err := net.Dial("tcp", reviewsFullAddress)
	if err != nil {
		fmt.Println("Reviews Conn error:", err)
		return
	}

	log.Printf("Games conn: %s", gamesFullAddress)
	log.Printf("Reviews conn: %s", reviewsFullAddress)

	wg.Add(2)

	go func() {
		<-c.sigChan
		c.handleSigterm()
	}()

	go func() {
		defer wg.Done()
		defer gamesConn.Close()

		readAndSendCSV(c.cfg.String("client.games_path", "data/games.csv"), uint8(message.GameIdMsg), gamesConn, &message.DataCSVGames{}, c)
		// header := make([]byte, 32)
		// if _, err = gamesConn.Read(header); err != nil {
		// 	logs.Logger.Errorf("Failed to read message: %v", err.Error())
		// }
	}()

	go func() {
		defer wg.Done()
		defer reviewsConn.Close()

		readAndSendCSV(c.cfg.String("client.reviews_path", "data/reviews.csv"), uint8(message.ReviewIdMsg), reviewsConn, &message.DataCSVReviews{}, c)
		// header := make([]byte, 32)
		// if _, err = reviewsConn.Read(header); err != nil {
		// 	logs.Logger.Errorf("Failed to read message: %v", err.Error())
		// }
	}()

	wg.Wait()

	logs.Logger.Info("Client exit...")

	c.Close(gamesConn, reviewsConn)
}

func (c *Client) handleSigterm() {
	logs.Logger.Info("Sigterm Signal Received... Shutting down")
	c.stoppedMutex.Lock()
	c.stopped = true
	c.stoppedMutex.Unlock()
}

func (c *Client) Close(gamesConn net.Conn, reviewsConn net.Conn) {
	gamesConn.Close()
	reviewsConn.Close()
	close(c.sigChan)
}

func (c *Client) startListener(wg *sync.WaitGroup, done chan bool) {
	defer wg.Done()

	listenAddress := c.cfg.String("client.address", "172.25.125.20")
	listenPort := c.cfg.String("client.port", "5050")
	fullListenAddress := listenAddress + ":" + listenPort

	listener, err := net.Listen("tcp", fullListenAddress)
	if err != nil {
		logs.Logger.Errorf("Error starting listener: %s", err)
		return
	}
	defer listener.Close()

	logs.Logger.Infof("Listening for results on: %s", fullListenAddress)

	for {
		select {
		case <-done:
			logs.Logger.Infof("All messages received, shutting down listener")
			return
		default:
			resultConn, err := listener.Accept()
			if err != nil {
				logs.Logger.Errorf("Error accepting connection: %s", err)
				continue
			}

			go handleResults(resultConn, done, wg)
		}
	}
}

func handleResults(conn net.Conn, done chan bool, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()

	buffer := make([]byte, 1024)
	messageCount := 0
	maxMessages := 5

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				logs.Logger.Infof("Connection closed")
			} else {
				logs.Logger.Errorf("Error reading from connection: %s", err)
			}
			return
		}

		if n > 0 {
			receivedData := string(buffer[:n])
			logs.Logger.Infof("Received: %s", receivedData)
			messageCount++
			if messageCount >= maxMessages {
				done <- true
				return
			}
		}
	}
}
