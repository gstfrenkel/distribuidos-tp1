package client

import (
	"fmt"
	"log"
	"net"
	"sync"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
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

	address := c.cfg.String("gateway_addr", "")

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to gateway:", err)
		return
	}

	log.Printf("Connected to: %s", address)

	defer conn.Close()

	var wg sync.WaitGroup

	wg.Add(2)

	// Debug read files
	// go readAndPrintCSV("data/games.csv", &DataCSVGames{}, &wg)
	// go readAndPrintCSV("data/reviews.csv", &DataCSVReviews{}, &wg)

	go readAndSendCSV("data/games.csv", uint8(message.GameIdMsg), conn, message.DataCSVGames{}, &wg)
	go readAndSendCSV("data/reviews.csv", uint8(message.ReviewIdMsg), conn, message.DataCSVReviews{}, &wg)

	wg.Wait()

	// TODO: recibir resultados del gateway
}
