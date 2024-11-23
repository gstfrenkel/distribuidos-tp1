package client

import (
	"encoding/binary"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

const LenFieldSize = 4
const ResultPrefixSize = 2

func (c *Client) startResultsListener(address string) {
	resultsPort := c.cfg.String("gateway.results_port", "5052")
	resultsFullAddress := address + ":" + resultsPort

	resultsConn, err := c.attemptConnection(resultsPort)
	if err != nil {
		logs.Logger.Errorf("Initial connection to %s failed: %v", resultsFullAddress, err)
		return
	}
	defer func() {
		if resultsConn != nil {
			resultsConn.Close()
		}
	}()

	logs.Logger.Infof("Connected to results on: %s", resultsFullAddress)

	messageCount := 0
	maxMessages := c.cfg.Int("client.max_messages", 5)
	receivedMap := make(map[string]bool)

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
			resultsConn = c.reconnect(resultsPort)
			continue
		}

		// Read message payload
		dataLen := binary.BigEndian.Uint32(lenBuffer)
		payload := make([]byte, dataLen)
		err = ioutils.ReadFull(resultsConn, payload, int(dataLen))
		if err != nil {
			logs.Logger.Errorf("Error reading payload: %v", err)
			resultsConn = c.reconnect(resultsPort)
			continue
		}

		receivedData := string(payload)
		prefix := receivedData[:ResultPrefixSize]
		if received, exists := receivedMap[prefix]; !exists || !received {
			if _, err := c.resultsFile.WriteString(receivedData + "\n\n"); err != nil {
				logs.Logger.Errorf("Error writing to results.txt: %v", err)
			}
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
