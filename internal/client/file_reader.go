package client

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"

	"tp1/pkg/logs"
	"tp1/pkg/message"
)

const (
	chunkSizeKey     = "client.chunk_size"
	chunkSizeDefault = 100
	timeoutKey       = "client.timeout"
	timeoutDefault   = 5
)

func (c *Client) readAndSendCSV(filename string, id uint8, conn net.Conn, dataStruct interface{}, address string) {
	file, reader, err := c.openAndPrepareCSVFile(filename)
	if err != nil {
		return
	}
	defer file.Close()

	timeout := c.cfg.Int(timeoutKey, timeoutDefault)
	batchSize := c.cfg.Int(chunkSizeKey, chunkSizeDefault)

	batchProcessor := &csvBatchProcessor{
		client:     c,
		file:       file,
		reader:     reader,
		conn:       conn,
		id:         id,
		dataStruct: dataStruct,
		batchSize:  batchSize,
		address:    address,
		timeout:    timeout,
	}

	batchProcessor.processCSVBatches()
}

type csvBatchProcessor struct {
	client     *Client
	file       *os.File
	reader     *csv.Reader
	conn       net.Conn
	id         uint8
	dataStruct interface{}
	batchSize  int
	address    string
	timeout    int
}

func (p *csvBatchProcessor) processCSVBatches() {
	var batchStartLine int
	currentBatch := uint32(0)

	for {
		currentLine, err := p.sendBatch(batchStartLine, currentBatch)
		if err == io.EOF {
			if err := p.handleFinalBatch(currentBatch); err != nil {
				continue
			}
			break
		}
		if err != nil {
			logs.Logger.Errorf("Error sending data: %s", err)
			p.rewindAndReconnect(batchStartLine)
			continue
		}

		if err := readAck(p.conn); err != nil {
			logs.Logger.Errorf("ACK error: %s", err)
			p.rewindAndReconnect(batchStartLine)
			continue
		}

		currentBatch++
		batchStartLine = currentLine
	}

	logs.Logger.Infof("Received EOF ACK for: %v", p.id)
}

func (p *csvBatchProcessor) sendBatch(startLine int, currentBatch uint32) (int, error) {
	lineCount := startLine
	for lineCount < startLine+p.batchSize {
		record, err := p.reader.Read()
		if err == io.EOF {
			return lineCount, err
		}
		if err != nil {
			return lineCount, fmt.Errorf("error reading CSV: %w", err)
		}

		if err := p.populateDataStruct(record); err != nil {
			logs.Logger.Errorf("Error populating data struct: %s", err)
			continue
		}

		if err := p.sendMessage(currentBatch); err != nil {
			return lineCount, err
		}

		lineCount++
	}
	return lineCount, nil
}

func (p *csvBatchProcessor) populateDataStruct(record []string) error {
	v := reflect.ValueOf(p.dataStruct).Elem()
	for i := 0; i < v.NumField(); i++ {
		if i < len(record) {
			field := v.Field(i)
			if err := setFieldValue(field, record[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			field.SetInt(intVal)
		}
	case reflect.Int:
		intVal, err := strconv.Atoi(value)
		if err == nil {
			field.SetInt(int64(intVal))
		}
	case reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err == nil {
			field.SetBool(boolVal)
		}
	default:
		logs.Logger.Infof("Unsupported type: %s", field.Kind())
	}
	return nil
}

func (p *csvBatchProcessor) sendMessage(currentBatch uint32) error {
	var dataBuf []byte
	var err error

	if p.id == uint8(message.ReviewIdMsg) {
		dataBuf, err = message.DataCSVReviews.ToBytes(*p.dataStruct.(*message.DataCSVReviews))
	} else {
		dataBuf, err = message.DataCSVGames.ToBytes(*p.dataStruct.(*message.DataCSVGames))
	}

	if err != nil {
		return fmt.Errorf("error encoding data: %w", err)
	}

	msg := message.ClientMessage{
		BatchNum: currentBatch,
		DataLen:  uint32(len(dataBuf)),
		Data:     dataBuf,
	}

	return message.SendMessage(p.conn, msg)
}

func (p *csvBatchProcessor) handleFinalBatch(currentBatch uint32) error {
	eofMsg := message.ClientMessage{
		BatchNum: currentBatch,
		DataLen:  0,
		Data:     nil,
	}

	if err := message.SendMessage(p.conn, eofMsg); err != nil {
		logs.Logger.Errorf("Error sending EOF message: %s", err)
		p.rewindAndReconnect(0)
		return err
	}

	if err := readAck(p.conn); err != nil {
		logs.Logger.Errorf("Error reading final ACK: %s", err)
		p.rewindAndReconnect(0)
		return err
	}

	return nil
}

func (p *csvBatchProcessor) rewindAndReconnect(batchStartLine int) {
	rewindReader(p.file, &p.reader, batchStartLine)
	p.conn = p.client.reconnect(p.address, p.timeout)
}

func (c *Client) openAndPrepareCSVFile(filename string) (*os.File, *csv.Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		logs.Logger.Errorf("Error opening CSV file: %s", err)
		return nil, nil, err
	}

	reader := csv.NewReader(file)

	if _, err = reader.Read(); err != nil {
		file.Close()
		if err == io.EOF {
			logs.Logger.Error("CSV file is empty.")
			return nil, nil, err
		}
		logs.Logger.Errorf("Error reading CSV file: %s", err)
		return nil, nil, err
	}

	return file, reader, nil
}

func rewindReader(file *os.File, reader **csv.Reader, batchStartLine int) {
	file.Seek(0, io.SeekStart)
	*reader = csv.NewReader(file)

	_, _ = (*reader).Read()

	for i := 0; i < batchStartLine; i++ {
		_, _ = (*reader).Read()
	}
}
