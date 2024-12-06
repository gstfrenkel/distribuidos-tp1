package main

import (
	"tp1/internal/healthcheck"
	f "tp1/internal/worker/filter"
	"tp1/pkg/logs"
)

func main() {
	hc, err := healthcheck.NewService()
	if err != nil {
		logs.Logger.Errorf("Failed to launch health check service: %s", err.Error())
		return
	}

	go hc.Listen()

	filter, err := f.NewReleaseDate()
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
