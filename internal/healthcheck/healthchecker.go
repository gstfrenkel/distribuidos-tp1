package healthcheck

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"tp1/pkg/config"
	"tp1/pkg/config/provider"
	"tp1/pkg/logs"
	"tp1/pkg/utils/io"
)

const (
	// Env vars keys
	hcIdKey       = "id"
	hcNextIdKey   = "next"
	hcNodesKey    = "nodes"
	hcNodesSepKey = " "

	// Config keys
	hcServerPort           = "hc.server-port"
	hcServerDefaultPort    = "9290"
	hcContainerNameKey     = "hc.container-name"
	hcDefaultContainerName = "healthchecker-%d"

	configFilePath   = "config.toml"
	sleepSecs        = 2
	waitAfterRestart = 5
	maxErrors        = 2
	hcMsg            = 1
	dockerRestart    = "docker restart "
	timeoutSecs      = 1
)

type HealthChecker struct {
	hcAddr     string
	serverPort string
	nextHc     string   //address of the next health checker
	nodes      []string //addresses of the nodes to check
	service    *Service
	finished   bool
	finishedMu sync.Mutex
}

func New() (*HealthChecker, error) {
	cfg, err := provider.LoadConfig(configFilePath)
	serverPort, containerName := getConfig(cfg)
	id, nextId, nodes, err := getEnvVars()
	if err != nil {
		return nil, err
	}

	hcService, err := NewService()
	if err != nil {
		return nil, err
	}

	hcAddr := hcAddrFromId(containerName, id)
	nextHc := hcAddrFromId(containerName, nextId)
	if nextId != id { //in case there is 1 hc, don't connect to itself
		nodes = append(nodes, nextHc)
	}

	return &HealthChecker{
		hcAddr:     hcAddr,
		nextHc:     nextHc,
		nodes:      nodes,
		serverPort: serverPort,
		service:    hcService,
	}, nil
}

// Start starts the health checker for every node
func (hc *HealthChecker) Start() {
	defer hc.service.Close()
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		hc.HandleSIGTERM()
	}()

	go func() {
		hc.service.Listen()
	}()

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

// Check checks if the node is alive and restarts it if it is not.
// nodeIp is the container name of the node
func (hc *HealthChecker) Check(nodeIp string) {
	for {
		hc.finishedMu.Lock()
		if hc.finished {
			hc.finishedMu.Unlock()
			return
		}
		hc.finishedMu.Unlock()

		nodeAddr := nodeIp + ":" + hc.serverPort
		conn, err := hc.connect(nodeAddr)
		if err != nil {
			logs.Logger.Errorf("Node conn error: %v", err)
			hc.restartNode(nodeIp)
			continue
		}

		errCount := hc.sendHcMsg(conn)
		conn.Close()

		if errCount == maxErrors {
			hc.restartNode(nodeIp)
		}
	}
}

// Send health check message to the node and wait for the ack.
// If it fails to send the message or receive ack maxError times, returns.
func (hc *HealthChecker) sendHcMsg(conn *net.UDPConn) int {
	errCount := 0
	buffer := make([]byte, msgBytes)
	for errCount < maxErrors {

		hc.finishedMu.Lock()
		if hc.finished {
			hc.finishedMu.Unlock()
			return errCount
		}
		hc.finishedMu.Unlock()

		_, err := conn.Write([]byte{hcMsg})
		if err != nil {
			errCount++
			logs.Logger.Errorf("Error sending health check message: %v. Count: %d", err, errCount)
			continue
		}

		err = conn.SetReadDeadline(time.Now().Add(timeoutSecs * time.Second))
		if err != nil {
			logs.Logger.Errorf("Error setting read deadline: %v", err)
		}

		_, err = conn.Read(buffer)
		if err != nil {
			errCount++
			logs.Logger.Debugf("Error recv health check ack: %v. Error count: %d", err, errCount)
		} else {
			errCount = 0
		}

		time.Sleep(waitAfterRestart * time.Second)
	}

	return errCount
}

// Connect to the node.
// If it fails to connect, it tries to reconnect maxErrors times.
// If it fails to reconnect, it restarts the node.
func (hc *HealthChecker) connect(nodeAddr string) (*net.UDPConn, error) {
	i := 0
	udpAddr, err := net.ResolveUDPAddr(transportProtocol, nodeAddr)
	if err != nil {
		logs.Logger.Errorf("Error resolving address %s: %v", nodeAddr, err)
		return nil, err
	}

	for i < maxErrors {
		conn, connErr := net.DialUDP(transportProtocol, nil, udpAddr)
		if connErr == nil {
			logs.Logger.Infof("Connected to node %v", udpAddr)
			return conn, nil
		}
		logs.Logger.Errorf("Error connecting to node %s: %v, retrying", nodeAddr, connErr)
		i++
		err = connErr
	}

	return nil, err
}

// Using DinD to restart the health checker
func (hc *HealthChecker) restartNode(containerName string) {
	logs.Logger.Errorf("Node %s is probably down", containerName)

	// Check if the container is running
	checkCmd := fmt.Sprintf("docker ps --filter name=%s --filter status=running", containerName)
	output, err := io.ExecCommand(checkCmd)
	if err != nil {
		logs.Logger.Errorf("Error checking container status: %s", err)
		return
	}

	if strings.Contains(output, containerName) {
		logs.Logger.Infof("Node %s is already running, won't restart it", containerName)
		return
	}

	// Restart the container
	_, err = io.ExecCommand(dockerRestart + containerName)
	if err != nil {
		logs.Logger.Errorf("Error restarting node: %s", err)
		return
	}

	logs.Logger.Infof("Node %s restarted", containerName)
	time.Sleep(1 * time.Second)
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

func hcAddrFromId(containerName string, id int) string {
	return fmt.Sprintf(containerName, id)
}

func (hc *HealthChecker) HandleSIGTERM() {
	logs.Logger.Info("Received SIGTERM, shutting down")
	hc.finishedMu.Lock()
	hc.finished = true
	hc.finishedMu.Unlock()
}
