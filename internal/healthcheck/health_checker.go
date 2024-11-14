package healthcheck

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

// Env vars keys
const hcIdKey = "id"
const hcNextIdKey = "next"
const hcNodesKey = "nodes"
const hcNodesSepKey = ","

// Config keys
const hcServerPort = "hc.server-port"
const hcServerDefaultPort = "9290"
const hcContainerNameKey = "hc.container-name"
const hcDefaultContainerName = "hc.health-checker-%d"

const configFilePath = "config.toml"
const sleepSecs = 10
const hcMsg = 1

type HealthChecker struct {
	id         uint8
	serverPort string
	nextHc     string   //address of the next health checker
	nodes      []string //addresses of the nodes to check
	service    *Service
}

func NewHc() (*HealthChecker, error) {
	cfg, err := provider.LoadConfig(configFilePath)
	serverPort, containerName := getConfig(cfg)
	id, nextId, nodes, err := getEnvVars()
	if err != nil {
		return nil, err
	}

	hcService, err := NewHcService()
	if err != nil {
		return nil, err
	}

	nextHc := fmt.Sprintf(containerName, nextId)

	return &HealthChecker{
		id:         uint8(id),
		nextHc:     nextHc,
		nodes:      append(nodes, nextHc),
		serverPort: serverPort,
		service:    hcService,
	}, nil
}

func getConfig(cfg config.Config) (string, string) {
	return cfg.String(hcServerPort, hcServerDefaultPort),
		cfg.String(hcContainerNameKey, hcDefaultContainerName)
}

func getEnvVars() (int, int, []string, error) {
	id, err := strconv.Atoi(os.Getenv(hcIdKey))
	nextId, err := strconv.Atoi(os.Getenv(hcNextIdKey))
	nodes := strings.Split(os.Getenv(hcNodesKey), hcNodesSepKey)

	return id, nextId, nodes, err
}

// Start starts the health checker for every node
func (hc *HealthChecker) Start() {
	wg := sync.WaitGroup{}
	wg.Add(len(hc.nodes))

	for _, node := range hc.nodes {
		go func(node string) {
			defer wg.Done()
			hc.Check(node)
		}(node)
	}

	wg.Wait()
}

func (hc *HealthChecker) Check(nodeIp string) {
	nodeAddr := nodeIp + ":" + hc.serverPort
	logs.Logger.Infof("GATEWAY %s", nodeAddr)
	conn, err := net.Dial(TransportProtocol, nodeAddr)
	if err != nil {
		logs.Logger.Errorf("Node conn error: %v", err)
		return
	}

	for {
		err := ioutils.SendAll(conn, []byte{hcMsg})
		logs.Logger.Infof("Sent health check to %s", nodeAddr)
		if err != nil { //TODO contar errores y reiniciar el nodo
			return
		}
		time.Sleep(sleepSecs * time.Second)
	}
}

/*// Using DinD to restart the health checker
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
}*/
