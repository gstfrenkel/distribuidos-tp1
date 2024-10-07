package main

import (
	"tp1/internal/worker/top_n_playtime_agg"
	"tp1/pkg/logs"
)

func main() {
	agg, err := top_n_playtime_agg.New() 
	if err != nil {
		logs.Logger.Errorf("Failed to create new results aggregator: %s", err.Error())
		return
	}

	if err = agg.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize new results aggregator: %s", err.Error())
		return
	}

	agg.Start()
}
