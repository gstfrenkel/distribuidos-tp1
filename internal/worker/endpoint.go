package worker

import "tp1/pkg/broker"

type Endpoint struct {
	consumers uint8
	outputs   []broker.EofDestination
}

func NewEndpoint(consumers uint8, outputs []broker.EofDestination) *Endpoint {
	return &Endpoint{consumers: consumers, outputs: outputs}
}
