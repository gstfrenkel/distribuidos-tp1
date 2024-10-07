package main

import (
	"tp1/internal/worker/platform_counter_agg"
	"tp1/pkg/logs"
)

func main() {
	filter, err := platform_counter_agg.New()
	if err != nil {
		logs.Logger.Errorf("Failed to create new games aggregator: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize new games aggregator: %s", err.Error())
		return
	}

	filter.Start()
}
