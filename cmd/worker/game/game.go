package main

import (
	"fmt"

	"tp1/internal/worker/game"
)

func main() {
	filter, err := game.New()
	if err != nil {
		fmt.Printf("Failed to create new games filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		fmt.Printf("Failed to initialize games filter: %s", err.Error())
		return
	}

	filter.Start()
}
