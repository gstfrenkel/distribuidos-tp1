package client

import (
	"fmt"
	"log"
	"net"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type Client struct {
	cfg config.Config
}

func New() (*Client, error) {

	config, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	return &Client{
		cfg: config,
	}, nil
}

func (c *Client) Start() {
	log.Println("Client running...")
	address := c.cfg.String("gateway.address", "127.0.0.1")
	port := c.cfg.String("gateway.port", "5050")
	fullAddress := address + ":" + port

	conn, err := net.Dial("tcp", fullAddress)

	if err != nil {
		fmt.Println("Error connecting to gateway:", err)
		return
	}

	log.Printf("test: %s", fullAddress)

	log.Printf("Connected to: %s", fullAddress)

	defer conn.Close()

	/*var wg sync.WaitGroup

	wg.Add(2)*/

	go readAndSendCSV(c.cfg.String("client.games_path", "data/games.csv"), uint8(message.GameIdMsg), conn, &message.DataCSVGames{})
	go readAndSendCSV(c.cfg.String("client.reviews_path", "data/reviews.csv"), uint8(message.ReviewIdMsg), conn, &message.DataCSVReviews{})

	/*go func() {
		defer wg.Done()
		readAndSendCSV(c.cfg.String("client.games_path", "data/games.csv"), uint8(message.GameIdMsg), conn, &message.DataCSVGames{})
	}()

	go func() {
		defer wg.Done()
		readAndSendCSV(c.cfg.String("client.reviews_path", "data/reviews.csv"), uint8(message.ReviewIdMsg), conn, &message.DataCSVReviews{})
	}()*/

	header := make([]byte, 1024)
	if _, err := conn.Read(header); err != nil {
		logs.Logger.Errorf("failed to read message header: %v", err.Error())
	}

	//wg.Wait()

	// TODO: recibir resultados del gateway
}
