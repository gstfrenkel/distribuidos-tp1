package main

import (
	"tp1/internal/worker/counter"
	"tp1/pkg/logs"
)

func main() {
	filter, err := counter.New()
	if err != nil {
		logs.Logger.Errorf("Failed to create new reviews filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize new reviews filter: %s", err.Error())
		return
	}

	filter.Start()
}
