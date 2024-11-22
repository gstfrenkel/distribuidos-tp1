package client

import (
	"encoding/binary"
	"net"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

const LenFieldSize = 4

// TODO: reconnect to gateway + Handle duplicate results
func (c *Client) startResultsListener(address string) {

	resultsPort := c.cfg.String("gateway.results_port", "5052")
	resultsFullAddress := address + ":" + resultsPort

	resultsConn, err := net.Dial("tcp", resultsFullAddress)
	if err != nil {
		logs.Logger.Errorf("Error connecting to results socket: %s", err)
		return
	}

	defer resultsConn.Close()

	err = c.sendClientID(resultsConn)
	if err != nil {
		logs.Logger.Errorf("Error sending client ID: %s", err)
		return
	}

	logs.Logger.Infof("Connected to results on: %s", resultsFullAddress)

	messageCount := 0
	maxMessages := c.cfg.Int("client.max_messages", 5)

	for {
		c.stoppedMutex.Lock()
		if c.stopped {
			c.stoppedMutex.Unlock()
			logs.Logger.Info("Shutting down results listener due to stopped signal.")
			return
		}
		c.stoppedMutex.Unlock()

		lenBuffer := make([]byte, LenFieldSize)
		err := ioutils.ReadFull(resultsConn, lenBuffer, LenFieldSize)
		if err != nil {
			logs.Logger.Errorf("Error reading length of message: %v", err)
			return
		}

		dataLen := binary.BigEndian.Uint32(lenBuffer)
		payload := make([]byte, dataLen)
		err = ioutils.ReadFull(resultsConn, payload, int(dataLen))
		if err != nil {
			logs.Logger.Errorf("Error reading payload from connection: %v", err)
			return
		}

		receivedData := string(payload)

		if _, err := c.resultsFile.WriteString(receivedData + "\n\n"); err != nil {
			logs.Logger.Errorf("Error writing to results.txt: %v", err)
		}

		messageCount++
		if messageCount >= maxMessages {
			return
		}
	}
}
