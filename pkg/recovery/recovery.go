package recovery

import (
	"io"

	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/sequence"
	ioutils "tp1/pkg/utils/io"
)

const filePath = "recovery.csv"

type Handler interface {
	Recover(ch chan<- Record)
	Log(record Record) error
	Close()
}

type handler struct {
	file *ioutils.File
}

func NewHandler() (Handler, error) {
	file, err := ioutils.NewFile(filePath)
	if err != nil {
		return nil, err
	}

	return &handler{
		file: file,
	}, nil
}

func (h *handler) Recover(ch chan<- Record) {
	defer close(ch)

	for {
		line, err := h.file.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
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
			[]byte(line[amqp.HeaderLen+dstSequenceHeaderLen+len(sequenceIds)]),
		)
	}
}

func (h *handler) Log(record Record) error {
	return h.file.Write(record.toString())
}

func (h *handler) Close() {
	h.file.Close()
}
