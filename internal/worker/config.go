package worker

import "tp1/pkg/broker"

type Config struct {
	Peers        int                  `json:"peers"`
	Query        any                  `json:"query"`
	InputQueues  []broker.Destination `json:"input-queues"`
	OutputQueues []broker.Destination `json:"output-queues"`
	Exchanges    []broker.Exchange    `json:"exchanges"`
	LogLevel     string               `json:"log_level"`
}
