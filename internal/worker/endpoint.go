package worker

import "tp1/pkg/broker"

type Endpoint struct {
	consumers uint8
	outputs   []broker.Route
}

func NewEndpoint(consumers uint8, outputs []broker.Route) *Endpoint {
	return &Endpoint{consumers: consumers, outputs: outputs}
}
