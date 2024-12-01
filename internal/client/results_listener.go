package client

import (
	"encoding/binary"
	"net"
	"tp1/pkg/logs"
	"tp1/pkg/utils/io"
)

const (
	LenFieldSize       = 4
	resultsPortKey     = "gateway.results_port"
	resultsPortDefault = "5052"
	transportProtocol  = "tcp"
	maxMsgKey          = "client.max_messages"
	maxMsgDefault      = 5
	ResultPrefixSize   = 2
)

func (c *Client) startResultsListener(address string) {
	resultsPort := c.cfg.String(resultsPortKey, resultsPortDefault)
	resultsFullAddress := address + ":" + resultsPort

	resultsConn, err := net.Dial(transportProtocol, resultsFullAddress)
	if err != nil {
		logs.Logger.Errorf("Initial connection to %s failed: %v", resultsFullAddress, err)
		return
	}

	err = c.sendClientID(resultsConn)
	if err != nil {
		logs.Logger.Errorf("Error sending client ID: %s", err)
		resultsConn.Close()

	}

	defer func() {
		if resultsConn != nil {
			resultsConn.Close()
		}
	}()

	logs.Logger.Infof("Connected to results on: %s", resultsFullAddress)

	messageCount := 0
	maxMessages := c.cfg.Int(maxMsgKey, maxMsgDefault)

	c.readResults(resultsConn, messageCount, maxMessages, resultsFullAddress)
}

func (c *Client) readResults(resultsConn net.Conn, messageCount int, maxMessages int, resultsFullAddress string) {
	timeout := c.cfg.Int(timeoutKey, timeoutDefault)
	receivedMap := make(map[string]bool)
	for {
		c.stoppedMutex.Lock()
		if c.stopped {
			c.stoppedMutex.Unlock()
			logs.Logger.Info("Shutting down results listener due to stopped signal.")
			return
		}
		c.stoppedMutex.Unlock()

		// Read payload size
		lenBuffer := make([]byte, LenFieldSize)
		err := io.ReadFull(resultsConn, lenBuffer, LenFieldSize)
		if err != nil {
			logs.Logger.Errorf("Error reading length of message: %v", err)
			resultsConn = c.reconnect(resultsFullAddress, timeout)
			continue
		}

		// Read message payload
		dataLen := binary.BigEndian.Uint32(lenBuffer)
		payload := make([]byte, dataLen)
		err = io.ReadFull(resultsConn, payload, int(dataLen))
		if err != nil {
			logs.Logger.Errorf("Error reading payload: %v", err)
			resultsConn = c.reconnect(resultsFullAddress, timeout)
			continue
		}

		receivedData := string(payload)
		prefix := receivedData[:ResultPrefixSize]
		if received, exists := receivedMap[prefix]; !exists || !received {
			c.writeDataToFile(receivedData)
			receivedMap[prefix] = true
			messageCount++
		} else {
			logs.Logger.Infof("Duplicate prefix %s received, skipping.", prefix)
		}

		if messageCount >= maxMessages {
			logs.Logger.Infof("Max messages (%d) reached. Exiting listener.", maxMessages)
			return
		}
	}
}

func (c *Client) writeDataToFile(receivedData string) {
	if _, err := c.resultsFile.WriteString(receivedData + "\n\n"); err != nil {
		logs.Logger.Errorf("Error writing to results.txt: %v", err)
	}
}
