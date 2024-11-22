package client

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"time"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

func readAndSendCSV(filename string, id uint8, conn net.Conn, dataStruct interface{}, c *Client, port string) {

	reconnect := func() net.Conn {
		var newConn net.Conn
		for {
			logs.Logger.Infof("Attempting to reconnect...")
			conn, err := c.reconnectToGateway(port)
			if err == nil {
				logs.Logger.Infof("Reconnected successfully.")
				newConn = conn
				break
			}
			logs.Logger.Errorf("Reconnect failed, retrying in 5 seconds: %v", err)
			time.Sleep(5 * time.Second)
		}
		return newConn
	}

	sendBatch := func(startLine int, batchSize int, reader *csv.Reader, dataStruct interface{}) (int, error) {
		lineCount := startLine
		for lineCount < startLine+batchSize {
			record, err := reader.Read()
			if err == io.EOF {
				return lineCount, err
			}
			if err != nil {
				return lineCount, fmt.Errorf("error reading CSV: %w", err)
			}

			// Populate the dataStruct with appropriate type conversion
			v := reflect.ValueOf(dataStruct).Elem()
			for i := 0; i < v.NumField(); i++ {
				if i < len(record) {
					field := v.Field(i)
					switch field.Kind() {
					case reflect.String:
						field.SetString(record[i])
					case reflect.Int64:
						if value, err := strconv.ParseInt(record[i], 10, 64); err == nil {
							field.SetInt(value)
						}
					case reflect.Int:
						if value, err := strconv.Atoi(record[i]); err == nil {
							field.SetInt(int64(value))
						}
					case reflect.Float64:
						if value, err := strconv.ParseFloat(record[i], 64); err == nil {
							field.SetFloat(value)
						}
					case reflect.Bool:
						if value, err := strconv.ParseBool(record[i]); err == nil {
							field.SetBool(value)
						}
					default:
						logs.Logger.Infof("Unsupported type: %s", field.Kind())
					}
				}
			}

			// Prepare data for sending
			var dataBuf []byte
			if id == uint8(message.ReviewIdMsg) {
				dataBuf, err = message.DataCSVReviews.ToBytes(*dataStruct.(*message.DataCSVReviews))
			} else {
				dataBuf, err = message.DataCSVGames.ToBytes(*dataStruct.(*message.DataCSVGames))
			}
			if err != nil {
				logs.Logger.Errorf("Error encoding data: %s", err)
				continue
			}

			msg := message.ClientMessage{
				DataLen: uint32(len(dataBuf)),
				Data:    dataBuf,
			}

			if err = message.SendMessage(conn, msg); err != nil {
				return lineCount, fmt.Errorf("error sending message: %w", err)
			}

			lineCount++
		}
		return lineCount, nil
	}

	var err error
	var file *os.File
	var reader *csv.Reader
	var batchStartLine, currentLine int
	const batchSize = 100 // TODO: read from config

	// Open CSV and initialize reader
	file, err = os.Open(filename)
	if err != nil {
		logs.Logger.Errorf("Error opening CSV file: %s", err)
		return
	}
	defer file.Close()

	reader = csv.NewReader(file)

	// Skip headers
	if _, err = reader.Read(); err != nil {
		if err == io.EOF {
			logs.Logger.Error("CSV file is empty.")
			return
		}
		logs.Logger.Errorf("Error reading CSV file: %s", err)
		return
	}

	for {
		currentLine, err = sendBatch(batchStartLine, batchSize, reader, dataStruct)
		if err == io.EOF {
			break
		}
		if err != nil {
			logs.Logger.Errorf("Error in sendBatch: %s", err)

			// Rewind reader and retry
			file.Seek(0, io.SeekStart)
			reader = csv.NewReader(file)
			for i := 0; i < batchStartLine; i++ {
				_, _ = reader.Read()
			}
			conn = reconnect()
			continue
		}

		// Read ACK for batch
		if err := readAck(conn); err != nil {
			logs.Logger.Errorf("ACK error: %s", err)

			// Rewind reader and retry
			file.Seek(0, io.SeekStart)
			reader = csv.NewReader(file)
			for i := 0; i < batchStartLine; i++ {
				_, _ = reader.Read()
			}
			conn = reconnect()
			continue
		}

		// Move to the next batch
		batchStartLine = currentLine
	}

	// Send EOF message
	for {
		eofMsg := message.ClientMessage{
			DataLen: 0,
			Data:    nil,
		}
		if err := message.SendMessage(conn, eofMsg); err != nil {
			logs.Logger.Errorf("Error sending EOF message: %s", err)
			// Reconnect and retry sending EOF
			conn = reconnect()
			continue
		}
		logs.Logger.Infof("Sent EOF for: %v", id)

		// Wait for EOF ACK
		if err := readAck(conn); err != nil {
			logs.Logger.Errorf("Error reading final ACK: %s", err)

			// Reconnect and retry reading the ACK
			conn = reconnect()
			continue
		}
		break
	}

	logs.Logger.Infof("Received EOF ACK for: %v", id)
}
