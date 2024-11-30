package client

import (
	"encoding/csv"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"

	"tp1/pkg/logs"
	"tp1/pkg/message"
)

func (c *Client) readAndSendCSV(filename string, id uint8, conn net.Conn, dataStruct interface{}) {
	err := c.sendClientID(conn)
	if err != nil {
		logs.Logger.Errorf("Error sending client ID: %s", err)
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		logs.Logger.Errorf("Error opening CSV file: %s", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read and ignore the first line (headers)
	if _, err = reader.Read(); err != nil {
		if err == io.EOF {
			logs.Logger.Error("CSV file is empty.")
			return
		}
		logs.Logger.Errorf("Error reading CSV file: %s", err)
		return
	}

	for {

		c.stoppedMutex.Lock()
		if c.stopped {
			c.stoppedMutex.Unlock()
			break
		}
		c.stoppedMutex.Unlock()

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logs.Logger.Errorf("Error reading CSV file: %s", err)
			return
		}

		// Populate the dataStruct with appropriate type conversion
		v := reflect.ValueOf(dataStruct).Elem()
		for i := 0; i < v.NumField(); i++ {
			if i < len(record) {
				field := v.Field(i)

				// Convert the string from the CSV to the correct type
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
			logs.Logger.Errorf("Error sending message: %s", err.Error())
			return
		}
	}

	// Send EOF message after breaking out of the loop
	eofMsg := message.ClientMessage{
		DataLen: 0, // DataLen = 0 for eof message.
		Data:    nil,
	}
	if err = message.SendMessage(conn, eofMsg); err != nil {
		logs.Logger.Errorf("Error sending EOF message: %s", err)
	}
	logs.Logger.Infof("Sent EOF for: %v", id)
}
