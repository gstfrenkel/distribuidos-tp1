package recovery

import (
	"io"
	"tp1/pkg/ioutils"
)

type Handler interface {
	Recover(ch chan<- []string)
	Log(record Record) error
	Close()
}

type handler struct {
	file *ioutils.File
}

func NewHandler() (Handler, error) {
	file, err := ioutils.NewFile()
	if err != nil {
		return nil, err
	}

	return &handler{
		file: file,
	}, nil
}

func (h *handler) Recover(ch chan<- []string) {
	defer close(ch)

	for {
		line, err := h.file.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				continue
			}
		}
		ch <- line
	}
}

func (h *handler) Log(record Record) error {
	return h.file.Write(record.toString())
}

func (h *handler) Close() {
	h.file.Close()
}
