package healthcheck

import (
	"fmt"
	"net"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
)

const (
	transportProtocol = "udp"
	msgBytes          = 1
	ackMsg            = 2
	portKey           = "service.port"
	defaultPort       = 9290
	maxErrKey         = "service.max-err"
	defaultMaxErr     = 3
	serviceCfgPath    = "healthcheck_service.toml"
)

type Service struct {
	listener *net.UDPConn
	maxErr   uint8
}

func NewService() (*Service, error) {
	cfg, err := provider.LoadConfig(serviceCfgPath)
	if err != nil {
		return nil, err
	}

	port := cfg.Uint16(portKey, defaultPort)
	maxError := cfg.Uint8(maxErrKey, defaultMaxErr)

	udpAddr, err := net.ResolveUDPAddr(transportProtocol, fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUDP(transportProtocol, udpAddr)
	if err != nil {
		return nil, err
	}

	return &Service{listener: listener, maxErr: maxError}, nil
}

// Listen reads the health checker messages.
// If it fails to read (timeout occurred), it means the health checker is down.
func (h *Service) Listen() {
	buf := make([]byte, msgBytes)
	i := uint8(0)
	for i < h.maxErr {
		_, addr, err := h.listener.ReadFromUDP(buf)
		if err != nil {
			logs.Logger.Warningf("Error reading health check message: %v", err)
			i++
			continue
		}

		_, err = h.listener.WriteToUDP([]byte{ackMsg}, addr)
		if err != nil {
			logs.Logger.Errorf("Error sending health check ack: %v", err)
		}
	}

	logs.Logger.Infof("Closing hc conn")
	h.Close()
}

func (h *Service) Close() {
	_ = h.listener.Close()
}
