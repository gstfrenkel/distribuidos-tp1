package main

import (
	"fmt"
	"log"

	"tp1/internal/worker/shooter"
)

func main() {
	filter, err := shooter.New()
	if err != nil {
		fmt.Printf("Failed to create new games filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		log.Printf("\n\n\nFailed to initialize games filter: %s\n\n\n", err.Error())
		fmt.Printf("\n\n\nFailed to initialize games filter: %s\n\n\n", err.Error())
		return
	}

	filter.Start()
}
