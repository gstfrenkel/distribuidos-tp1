package client

import (
	"encoding/binary"
	"net"
	"tp1/pkg/ioutils"
	"tp1/pkg/logs"
)

const (
	LenFieldSize       = 4
	resultsPortKey     = "gateway.results_port"
	resultsPortDefault = "5052"
	transportProtocol  = "tcp"
	maxMsgKey          = "client.max_messages"
	maxMsgDefault      = 5
)

func (c *Client) startResultsListener(address string) {
	resultsPort := c.cfg.String(resultsPortKey, resultsPortDefault)
	resultsFullAddress := address + ":" + resultsPort

	resultsConn, err := net.Dial(transportProtocol, resultsFullAddress)
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
	maxMessages := c.cfg.Int(maxMsgKey, maxMsgDefault)

	c.readResults(resultsConn, messageCount, maxMessages, resultsPort)
}

func (c *Client) readResults(resultsConn net.Conn, messageCount int, maxMessages int, resultsPort string) {

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

		c.writeDataToFile(receivedData)

		messageCount++
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
