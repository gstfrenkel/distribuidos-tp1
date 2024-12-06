package recovery

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/sequence"

	ioutils "tp1/pkg/utils/io"
)

const (
	dirPath   = "recovery/"
	extension = ".csv"
)

type Handler struct {
	files map[string]*ioutils.File
}

// NewHandler creates a new recovery handler.
func NewHandler() (*Handler, error) {
	files := make(map[string]*ioutils.File)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != extension {
			continue
		}

		file, err := ioutils.NewFile(entry.Name())
		if err != nil {
			return nil, err
		}

		files[strings.TrimSuffix(entry.Name(), extension)] = file
	}

	return &Handler{
		files: files,
	}, nil
}

// Recover reads each line of the underlying file, parses it, and sends it through a Record channel.
// Upon failure when reading or parsing, the line gets skipped.
func (h *Handler) Recover(ch chan<- Record) {
	defer close(ch)

	var wg sync.WaitGroup
	wg.Add(len(h.files))

	for _, file := range h.files {
		go recoverFile(file, ch, &wg)
	}

	wg.Wait()
}

// Log saves a record into the underlying file.
func (h *Handler) Log(record Record) error {
	var err error

	file, ok := h.files[record.Header().ClientId]
	if !ok {
		file, err = ioutils.NewFile(fullPath(record.Header().ClientId))
		if err != nil {
			return err
		}
	}

	h.files[record.Header().ClientId] = file
	return file.Write(record.toString())
}

func (h *Handler) Delete(clientId string) {
	file, ok := h.files[clientId]
	if !ok {
		return
	}
	file.Close()
	if err := os.Remove(fullPath(clientId)); err != nil {
		logs.Logger.Errorf("failed to delete file: %s", err.Error())
	}
	delete(h.files, clientId)
}

// Close closes the files descriptors linked to the underlying files.
func (h *Handler) Close() {
	for _, file := range h.files {
		file.Close()
	}
}

func recoverFile(file *ioutils.File, ch chan<- Record, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		line, err := file.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logs.Logger.Errorf("Failed to read line: %v", err)
			continue
		}

		header, err := amqp.HeaderFromStrings(line)
		if err != nil {
			logs.Logger.Errorf("failed to recover header: %s", err.Error())
			continue
		}

		sequenceIds, err := sequence.DstsFromStrings(line[amqp.HeaderLen : len(line)-1])
		if err != nil {
			logs.Logger.Errorf("failed to recover sequence: %s", err.Error())
			continue
		}

		ch <- NewRecord(
			*header,
			sequenceIds,
			[]byte(line[len(line)-1]),
		)
	}
}

func fullPath(clientId string) string {
	return dirPath + clientId + extension
}
