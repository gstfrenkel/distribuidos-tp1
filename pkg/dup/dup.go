package dup

import (
	"tp1/pkg/sequence"
)

type Handler interface {
	Add(sequence.Source)
	IsDuplicate(sequence.Source) bool
}

type handler struct {
	dupsByWorker map[uint8]uint64
}

func NewHandler() Handler {
	return &handler{dupsByWorker: make(map[uint8]uint64)}
}

func (h *handler) Add(seq sequence.Source) {
	h.dupsByWorker[seq.WorkerId()] = seq.Id() + 1
}

func (h *handler) IsDuplicate(seq sequence.Source) bool {
	nextSequenceId := h.dupsByWorker[seq.WorkerId()]
	if seq.Id() < nextSequenceId {
		return true
	}

	h.dupsByWorker[seq.WorkerId()] = seq.Id() + 1

	return false
}
