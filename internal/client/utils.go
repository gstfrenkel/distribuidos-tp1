package client

import (
	"fmt"
	"net"
	"os"
	"tp1/internal/gateway/id_generator"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

func openResultsFile(c *Client) error {
	fileName := fmt.Sprintf("/app/data/results_%s.txt", c.clientId)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		logs.Logger.Errorf("Error opening results file: %v", err)
		return err
	}

	c.resultsFile = file
	return nil
}

func fetchClientID(c *Client, address string) error {
	idsPort := c.cfg.String("gateway.ids_port", "5053")
	idsFullAddress := address + ":" + idsPort
	idConn, err := net.Dial("tcp", idsFullAddress)
	if err != nil {
		logs.Logger.Errorf("ID connection error: %v", err)
		return err
	}
	defer idConn.Close()

	clientIdBuffer := make([]byte, id_generator.ClientIdLen)
	if err := ioutils.ReadFull(idConn, clientIdBuffer, id_generator.ClientIdLen); err != nil {
		logs.Logger.Errorf("Error reading client ID: %v", err)
		return err
	}
	c.clientId = id_generator.DecodeClientId(clientIdBuffer)
	logs.Logger.Infof("Assigned client ID: %s", c.clientId)
	return nil
}

func (c *Client) handleSigterm() {
	logs.Logger.Info("Sigterm Signal Received... Shutting down")
	c.stoppedMutex.Lock()
	c.stopped = true
	c.stoppedMutex.Unlock()
}

func (c *Client) Close(gamesConn net.Conn, reviewsConn net.Conn) {
	gamesConn.Close()
	reviewsConn.Close()
	c.resultsFile.Close()
	close(c.sigChan)
}

func sendClientID(c *Client, conn net.Conn) error {
	clientIdBuf := id_generator.EncodeClientId(c.clientId)
	if err := ioutils.SendAll(conn, clientIdBuf); err != nil {
		logs.Logger.Errorf("Error sending client ID: %s", err)
		return err
	}
	logs.Logger.Infof("Client ID %d sent", c.clientId)
	return nil
}
