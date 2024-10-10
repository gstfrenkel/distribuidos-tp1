package main

import (
	"log"
	"tp1/internal/client"
)

func main() {

	client, err := client.New()

	if err != nil {
		log.Printf("Failed to launch client: %s", err.Error())
		return
	}

	client.Start()
}
