package top_n

import (
	"fmt"
	"tp1/internal/worker/top_n"
)

func main() {
	topNworker, err := top_n.New()
	if err != nil {
		fmt.Printf("Failed to create new top 5 positive reviews worker: %s", err.Error())
		return
	}

	if err = topNworker.Init(); err != nil {
		fmt.Printf("Failed to initialize top 5 positive reviews worker: %s", err.Error())
		return
	}

	topNworker.Start()
}
