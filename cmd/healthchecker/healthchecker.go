package main

import (
	"log"
	"tp1/internal/healthcheck"
)

func main() {

	hc, err := healthcheck.New()

	if err != nil {
		log.Printf("Failed to launch healthchecker: %s", err.Error())
		return
	}

	hc.Start()
}
