package healthcheck

import (
	"fmt"
	"net"
	"tp1/pkg/logs"
)

const port = 9290
const transportProtocol = "udp"
const msgBytes = 1
const ackMsg = 2

type Service struct {
	listener *net.UDPConn
}

func NewHcService() (*Service, error) {
	udpAddr, err := net.ResolveUDPAddr(transportProtocol, fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUDP(transportProtocol, udpAddr)
	if err != nil {
		return nil, err
	}
	logs.Logger.Infof("Health check service listening on port %d", port)

	return &Service{listener: listener}, nil
}

// Listen reads the health checker messages.
// If it fails to read (timeout occurred), it means the health checker is down.
func (h *Service) Listen() {
	buf := make([]byte, msgBytes)
	for { //todo: check sigterm to stop?
		_, addr, err := h.listener.ReadFromUDP(buf)
		if err != nil {
			logs.Logger.Errorf("Error reading health check message: %v", err)
			continue
		}

		logs.Logger.Infof("Received health check message from: %s", addr)

		_, err = h.listener.WriteToUDP([]byte{ackMsg}, addr)
		if err != nil {
			logs.Logger.Errorf("Error sending health check ack: %v", err)
		}
	}
}

func (h *Service) Close() error {
	return h.listener.Close()
}
