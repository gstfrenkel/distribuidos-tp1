package client

import (
	"log"
	"os"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
)

func initializeConfig() (config.Config, error) {
	cfg, err := provider.LoadConfig("config.toml")
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func initializeLog() {
	logFile, err := os.OpenFile("client.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	log.Println("Logging initialized")
}

func main() {
	cfg, err := initializeConfig()
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	initializeLog()
	log.Printf("Client configuration: %+v", cfg)

	client := NewClient()
	client.Start()
}
