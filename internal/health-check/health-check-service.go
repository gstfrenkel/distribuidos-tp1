package health_check

import (
	"net"
	"strconv"
	"time"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

const port = 9290
const transportProtocol = "tcp"
const timeoutSec = 10
const msgBytes = 1

type HealthCheckService struct {
	listener          net.Listener
	healthCheckerDown bool
}

func New() (*HealthCheckService, error) {
	listener, err := net.Listen(transportProtocol, ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	return &HealthCheckService{listener: listener, healthCheckerDown: false}, nil
}

// Listen listens for incoming health checker connections.
// There should be only ONE health checker trying to connect to a node, that's why we don't need to handle multiple connections.
func (h *HealthCheckService) Listen() error {
	for {
		if h.healthCheckerDown {
			restartHealthChecker()
			h.healthCheckerDown = false
		}

		conn, err := h.listener.Accept()
		if err != nil {
			return err
		}

		h.handleHealthChecker(conn)
	}
}

// handleHealthChecker reads the health checker messages.
// If it fails to read (timeout occurred), it means the health checker is down.
func (h *HealthCheckService) handleHealthChecker(conn net.Conn) {
	buf := make([]byte, msgBytes)
	for { //todo: check sigterm to stop
		_ = conn.SetReadDeadline(time.Now().Add(timeoutSec * time.Second))
		err := ioutils.ReadFull(conn, buf, msgBytes)
		if err != nil {
			logs.Logger.Errorf("Health checker down: %s", err)
			h.healthCheckerDown = true
			_ = conn.Close()
			return
		}
	}
}

// Using DinD to restart the health checker
// TODO sincronizar esto para que no lo intenten reiniciar todos los nodos al mismo tiempo
func restartHealthChecker() {
	logs.Logger.Info("Restarting health checker")
	err := ioutils.ExecCommand("docker stop health-checker")
	if err != nil {
		logs.Logger.Errorf("Error stopping health checker: %s", err)
	}

	err = ioutils.ExecCommand("docker start health-checker")
	if err != nil {
		logs.Logger.Errorf("Error restarting health checker: %s", err)
	}
}

func (h *HealthCheckService) Close() error {
	return h.listener.Close()
}
