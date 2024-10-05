package main

import (
	"log"
	"tp1/internal/worker/indie"
)

func main() {
	filter, err := indie.New()
	if err != nil {
		fmt.Printf("Failed to create new games filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		log.Printf("\n\n\nFailed to initialize games filter: %s\n\n\n", err.Error())
		return
	}

	filter.Start()
}
