package main

import (
	"tp1/internal/worker/action"
	"tp1/pkg/logs"
)

func main() {
	filter, err := action.New()
	if err != nil {
		logs.Logger.Errorf("Failed to create new games filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize new games filter: %s", err.Error())
		return
	}

	filter.Start()
}
