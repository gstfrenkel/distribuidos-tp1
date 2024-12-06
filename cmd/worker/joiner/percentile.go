package main

import (
	"tp1/internal/healthcheck"
	"tp1/internal/worker/joiner"
	"tp1/pkg/logs"
)

func main() {
	hc, err := healthcheck.NewService()
	if err != nil {
		logs.Logger.Errorf("Failed to launch health checker: %s", err.Error())
		return
	}

	go hc.Listen()

	filter, err := joiner.NewPercentile()
	if err != nil {
		logs.Logger.Errorf("Failed to create new joiner: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize new joiner: %s", err.Error())
		return
	}

	filter.Start()
}
