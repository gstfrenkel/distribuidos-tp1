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
	logs.Logger.Info("Client running...")
	address := c.cfg.String("gateway.address", "172.25.125.100")

	gamesPort := c.cfg.String("gateway.games_port", "5050")
	gamesFullAddress := address + ":" + gamesPort

	reviewsPort := c.cfg.String("gateway.reviews_port", "5051")
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

	defer gamesConn.Close()
	defer reviewsConn.Close()

	go readAndSendCSV(c.cfg.String("client.games_path", "data/games.csv"), uint8(message.GameIdMsg), gamesConn, &message.DataCSVGames{})
	go readAndSendCSV(c.cfg.String("client.reviews_path", "data/reviews.csv"), uint8(message.ReviewIdMsg), reviewsConn, &message.DataCSVReviews{})

	header := make([]byte, 1024)
	if _, err = gamesConn.Read(header); err != nil {
		logs.Logger.Errorf("failed to read message: %v", err.Error())
	}

	if _, err = reviewsConn.Read(header); err != nil {
		logs.Logger.Errorf("failed to read message: %v", err.Error())
	}

	// TODO: recibir resultados del gateway
}
