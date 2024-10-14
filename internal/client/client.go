package client

import (
	"encoding/binary"
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
	resultsFile  *os.File
}

func New() (*Client, error) {
	config, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile("/app/data/results.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
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
		resultsFile:  file,
	}, nil
}

func (c *Client) Start() {
	logs.Logger.Info("Client running...")

	wg := sync.WaitGroup{}
	done := make(chan bool)

	wg.Add(1)
	go func() {
		c.startResultsListener(&wg, done)
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
	c.resultsFile.Close()
	close(c.sigChan)
}

func (c *Client) startResultsListener(wg *sync.WaitGroup, done chan bool) {
	defer wg.Done()

	addr := c.cfg.String("gateway.address", "172.25.125.20")
	resultsPort := c.cfg.String("gateway.results_port", "5052")
	resultsFullAddress := addr + ":" + resultsPort

	resultsConn, err := net.Dial("tcp", resultsFullAddress)
	if err != nil {
		logs.Logger.Errorf("Error connecting to results socket: %s", err)
		return
	}

	defer resultsConn.Close()

	logs.Logger.Infof("Connected to results on: %s", resultsFullAddress)

	messageCount := 0
	maxMessages := 5

	for {
		c.stoppedMutex.Lock()
		if c.stopped {
			c.stoppedMutex.Unlock()
			logs.Logger.Info("Shutting down results listener due to stopped signal.")
			return
		}
		c.stoppedMutex.Unlock()

		lenBuffer := make([]byte, 4)
		err := readFull(resultsConn, lenBuffer, 4)
		if err != nil {
			logs.Logger.Errorf("Error reading length of message: %v", err)
			return
		}

		dataLen := binary.BigEndian.Uint32(lenBuffer)
		payload := make([]byte, dataLen)
		err = readFull(resultsConn, payload, int(dataLen))
		if err != nil {
			logs.Logger.Errorf("Error reading payload from connection: %v", err)
			return
		}

		receivedData := string(payload)
		logs.Logger.Infof("Received: %s", receivedData)

		if _, err := c.resultsFile.WriteString(receivedData + "\n\n"); err != nil {
			logs.Logger.Errorf("Error writing to results.txt: %v", err)
		}

		messageCount++
		if messageCount >= maxMessages {
			done <- true
			return
		}
	}
}

func readFull(conn net.Conn, buffer []byte, n int) error {
	totalBytesRead := 0

	for totalBytesRead < n {
		bytesRead, err := conn.Read(buffer[totalBytesRead:])
		if err != nil {
			return err
		}
		if bytesRead == 0 {
			return io.EOF
		}
		totalBytesRead += bytesRead
	}

	return nil
}
