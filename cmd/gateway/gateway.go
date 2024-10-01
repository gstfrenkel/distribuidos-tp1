package main

import (
	"tp1/internal/gateway"
	"tp1/pkg/logs"
)

var log, _ = logs.GetLogger("gateway")

func main() {
	g, err := gateway.New()
	if err != nil {
		return
	}
	err = logs.InitLogger(g.Config.String("gateway.log_level", "INFO"))
	if err != nil {
		log.Errorf("Failed to create new gateway: %s", err.Error())
		return
	}
	log.Infof("Starting gateway..........")

	g.Start()
	//g.End()
}
