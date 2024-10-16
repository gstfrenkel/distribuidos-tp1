package client

import (
	"encoding/binary"
	"io"
	"net"
	"tp1/pkg/logs"
)

const LenFieldSize = 4

func (c *Client) startResultsListener(done chan bool, address string) {

	resultsPort := c.cfg.String("gateway.results_port", "5052")
	resultsFullAddress := address + ":" + resultsPort

	resultsConn, err := net.Dial("tcp", resultsFullAddress)
	if err != nil {
		logs.Logger.Errorf("Error connecting to results socket: %s", err)
		return
	}

	defer resultsConn.Close()

	logs.Logger.Infof("Connected to results on: %s", resultsFullAddress)

	messageCount := 0
	maxMessages := 5

	for {
		c.stoppedMutex.Lock()
		if c.stopped {
			c.stoppedMutex.Unlock()
			logs.Logger.Info("Shutting down results listener due to stopped signal.")
			return
		}
		c.stoppedMutex.Unlock()

		lenBuffer := make([]byte, LenFieldSize)
		err := readFull(resultsConn, lenBuffer, LenFieldSize)
		if err != nil {
			logs.Logger.Errorf("Error reading length of message: %v", err)
			return
		}

		dataLen := binary.BigEndian.Uint32(lenBuffer)
		payload := make([]byte, dataLen)
		err = readFull(resultsConn, payload, int(dataLen))
		if err != nil {
			logs.Logger.Errorf("Error reading payload from connection: %v", err)
			return
		}

		receivedData := string(payload)
		logs.Logger.Infof("Received: %s", receivedData)

		if _, err := c.resultsFile.WriteString(receivedData + "\n\n"); err != nil {
			logs.Logger.Errorf("Error writing to results.txt: %v", err)
		}

		messageCount++
		if messageCount >= maxMessages {
			done <- true
			return
		}
	}
}

func readFull(conn net.Conn, buffer []byte, n int) error {
	totalBytesRead := 0

	for totalBytesRead < n {
		bytesRead, err := conn.Read(buffer[totalBytesRead:])
		if err != nil {
			return err
		}
		if bytesRead == 0 {
			return io.EOF
		}
		totalBytesRead += bytesRead
	}

	return nil
}
