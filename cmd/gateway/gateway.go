package main

import (
	"log"
	"tp1/internal/gateway"
)

func main() {
	g, err := gateway.New()
	if err != nil {
		log.Printf("Failed to create new gateway: %s", err.Error())
		return
	}

	g.Start()
	g.End()
}
