package client

import (
	"log"
)

type Client struct {
	// TODO
}

// NewClient creates a new Client instance with the provided configuration parameters.
func NewClient() *Client {
	return &Client{}
}

// Start launches the client and begins processing queues and files.
func (c *Client) Start() {
	log.Println("Client started")

	// Todo
}
