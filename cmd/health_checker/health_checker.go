package main

import (
	"log"
	"tp1/internal/health_check"
)

func main() {

	hc, err := health_check.NewHc()

	if err != nil {
		log.Printf("Failed to launch health-checker: %s", err.Error())
		return
	}

	hc.Start()
}
