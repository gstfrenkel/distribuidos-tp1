package main

import (
	"fmt"

	"tp1/internal/worker/review"
)

func main() {
	filter, err := review.New()
	if err != nil {
		fmt.Printf("Failed to create new reviews filter: %s", err.Error())
		return
	}

	filter.Start()
}
