package main

import (
	"os"

	"tp1/internal/worker/joiner"
	"tp1/pkg/logs"
)

func main() {
	filter, err := joiner.NewCounter()
	if err != nil {
		logs.Logger.Errorf("Failed to create new joiner: %s", err.Error())
		os.Exit(1)
	}

	if err = filter.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize new joiner: %s", err.Error())
		os.Exit(1)
	}

	filter.Start()
}
