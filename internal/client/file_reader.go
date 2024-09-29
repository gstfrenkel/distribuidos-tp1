package client

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
)

func readAndSendCSV(filename string, id uint8, conn net.Conn, dataStruct interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read and ignore the first line (headers)
	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			fmt.Println("CSV file is empty.")
			return
		}
		fmt.Println("Error reading CSV file:", err)
		return
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading CSV file:", err)
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
				}
			}
		}

		// Prepare data for sending
		var dataBuf bytes.Buffer
		encoder := gob.NewEncoder(&dataBuf)
		if err := encoder.Encode(dataStruct); err != nil {
			fmt.Println("Error encoding data:", err)
			continue
		}

		msg := Message{
			ID:      id,
			DataLen: uint64(dataBuf.Len()),
			Data:    dataBuf.Bytes(),
		}

		if err := sendMessage(conn, msg); err != nil {
			fmt.Println("Error sending message:", err)
		}

	}

	// Send EOF message
	eofMsg := Message{
		ID:      3,
		DataLen: 0,
		Data:    nil,
	}
	if err := sendMessage(conn, eofMsg); err != nil {
		fmt.Println("Error sending EOF message:", err)
	}
}

// Debug function
func readAndPrintCSV(filename string, dataStruct interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read and ignore the first line (headers)
	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			fmt.Println("CSV file is empty.")
			return
		}
		fmt.Println("Error reading CSV file:", err)
		return
	}

	// Counter for the number of records printed
	printCount := 0
	const maxPrintCount = 10

	for {
		if printCount >= maxPrintCount {
			break
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading CSV file:", err)
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
				}
			}
		}

		// Print the populated struct
		fmt.Printf("Record %d: %+v\n", printCount+1, dataStruct)

		// Increment print count
		printCount++
	}
}
