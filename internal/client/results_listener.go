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
	maxMessages := c.cfg.Int(maxMsgKey, maxMsgDefault)

	c.readResults(resultsConn, messageCount, maxMessages)
}

func (c *Client) readResults(resultsConn net.Conn, messageCount int, maxMessages int) {
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

		c.writeDataToFile(receivedData)

		messageCount++
		if messageCount >= maxMessages {
			return
		}
	}
}

func (c *Client) writeDataToFile(receivedData string) {
	if _, err := c.resultsFile.WriteString(receivedData + "\n\n"); err != nil {
		logs.Logger.Errorf("Error writing to results.txt: %v", err)
	}
}
