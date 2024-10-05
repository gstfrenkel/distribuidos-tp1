package main

import (
	"log"

	"tp1/internal/worker/action"
)

func main() {
	filter, err := action.New()
	if err != nil {
		log.Println("Failed to create new games filter")
		return
	}

	if err = filter.Init(); err != nil {
		log.Printf("\n\n\nFailed to initialize games filter: %s\n\n\n", err.Error())
		return
	}

	filter.Start()
}
