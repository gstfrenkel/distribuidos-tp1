package worker

import "tp1/pkg/broker"

type Endpoint struct {
	consumers uint8
	outputs   []broker.Destination
}

func NewEndpoint(consumers uint8, outputs []broker.Destination) *Endpoint {
	return &Endpoint{consumers: consumers, outputs: outputs}
}
