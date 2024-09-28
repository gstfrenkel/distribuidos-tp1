package client

import (
	"log"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
)

type Client struct {
	Config config.Config
}

func New() (*Client, error) {

	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}

	return &Client{
		Config: cfg,
	}, nil
}

func (c *Client) Start() {
	log.Println("Client running...")

}
