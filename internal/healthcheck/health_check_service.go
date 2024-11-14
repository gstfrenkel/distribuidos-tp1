package healthcheck

import (
	"net"
	"strconv"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

const port = 9290
const TransportProtocol = "tcp"
const msgBytes = 1

type Service struct {
	listener net.Listener
}

func NewHcService() (*Service, error) {
	listener, err := net.Listen(TransportProtocol, ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	return &Service{listener: listener}, nil
}

// Listen listens for incoming health checker connections.
// There should be only ONE health checker trying to connect to a node, that's why we don't need to handle multiple connections.
func (h *Service) Listen() error {
	logs.Logger.Infof("Health checker listening on: %s", h.listener.Addr().String())
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			return err
		}
		h.handleHealthChecker(conn)
	}
}

// handleHealthChecker reads the health checker messages.
// If it fails to read (timeout occurred), it means the health checker is down.
func (h *Service) handleHealthChecker(conn net.Conn) {
	buf := make([]byte, msgBytes)
	for { //todo: check sigterm to stop
		err := ioutils.ReadFull(conn, buf, msgBytes)
		logs.Logger.Infof("Health checker message: %s", buf)
		if err != nil {
			logs.Logger.Errorf("Health checker down: %s", err)
			_ = conn.Close()
			return
		}
	}
}

func (h *Service) Close() error {
	return h.listener.Close()
}
