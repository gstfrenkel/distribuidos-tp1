package main

import (
	"tp1/internal/worker/platform"
	"tp1/pkg/logs"
)

func main() {
	filter, err := platform.New()
	if err != nil {
		logs.Logger.Printf("Failed to create new games filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		logs.Logger.Printf("\n\n\nFailed to initialize games filter: %s\n\n\n", err.Error())
		return
	}

	filter.Start()
}
