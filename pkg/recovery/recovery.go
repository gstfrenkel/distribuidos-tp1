package recovery

import (
	"tp1/pkg/ioutils"
)

type Handler interface {
	RecoverNextLine(parse func([]string) (any, error)) (any, error)
	Log(record []string) error
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

func (h *handler) RecoverNextLine(parse func([]string) (any, error)) (any, error) {
	line, err := h.file.Read()
	if err != nil {
		return nil, err
	}

	return parse(line)
}

func (h *handler) Log(record []string) error {
	return h.file.Write(record)
}

func (h *handler) Close() {
	h.file.Close()
}
