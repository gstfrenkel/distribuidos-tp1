package main

import (
	"fmt"
	"tp1/internal/worker/top_n"
)

func main() {
	topNworker, err := top_n.New()
	if err != nil {
		fmt.Printf("Failed to create new top N worker: %s", err.Error())
		return
	}

	if err = topNworker.Init(); err != nil {
		fmt.Printf("Failed to initialize top N worker: %s", err.Error())
		return
	}

	topNworker.Start()
}
