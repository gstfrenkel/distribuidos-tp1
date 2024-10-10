package main

import (
	"tp1/internal/gateway"
	"tp1/pkg/logs"
)

func main() {
	g, err := gateway.New()
	if err != nil {
		logs.Logger.Errorf("Failed to create new gateway: %s", err.Error())
		return
	}
	err = logs.InitLogger(g.Config.String("gateway.log_level", "INFO"))
	if err != nil {
		logs.Logger.Errorf("Failed to create new gateway: %s", err.Error())
		return
	}

	g.Start()
}
