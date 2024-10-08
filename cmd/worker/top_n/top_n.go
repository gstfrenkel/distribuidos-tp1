package main

import (
	"tp1/internal/worker/top_n"
	"tp1/pkg/logs"
)

func main() {
	topNworker, err := top_n.New()
	if err != nil {
		logs.Logger.Infof("Failed to create new top N worker: %s", err.Error())
		return
	}

	if err = topNworker.Init(); err != nil {
		logs.Logger.Infof("Failed to initialize top N worker: %s", err.Error())
		return
	}

	topNworker.Start()
}
