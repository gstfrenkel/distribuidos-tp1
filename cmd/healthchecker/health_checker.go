package main

import (
	"log"
	"tp1/internal/healthcheck"
)

func main() {

	hc, err := healthcheck.NewHc()

	if err != nil {
		log.Printf("Failed to launch healthchecker: %s", err.Error())
		return
	}

	hc.Start()
}
