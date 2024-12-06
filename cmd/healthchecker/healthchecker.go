package main

import (
	"log"
	"tp1/internal/healthcheck"
	"tp1/pkg/logs"
)

func main() {
	service, err := healthcheck.NewService()
	if err != nil {
		logs.Logger.Errorf("Failed to launch health check service: %s", err.Error())
		return
	}

	go service.Listen()

	hc, err := healthcheck.New()
	if err != nil {
		log.Printf("Failed to launch healthchecker: %s", err.Error())
		return
	}

	hc.Start()
}
