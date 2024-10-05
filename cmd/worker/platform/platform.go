package main

import (
	"fmt"

	"tp1/internal/worker/platform"
)

func main() {
	filter, err := platform.New()
	if err != nil {
		fmt.Printf("Failed to create new games filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		fmt.Printf("\n\n\nFailed to initialize games filter: %s\n\n\n", err.Error())
		return
	}

	filter.Start()
}
