package dup

import (
	"tp1/pkg/sequence"
)

type key struct {
	workerUuid string
	clientId   string
}

func newKey(workerUuid, clientId string) key {
	return key{workerUuid: workerUuid, clientId: clientId}
}

// Handler is an implementation of the Handler interface.
// It tracks the next expected sequence ID for each worker using a map.
type Handler struct {
	dupsByWorker map[key]uint64 // dupsByWorker saves the next expected sequence ID by worker UUID.
}

// NewHandler creates and returns a new instance of a Handler.
// It initializes the internal map to track sequence IDs for workers.
func NewHandler() *Handler {
	return &Handler{dupsByWorker: make(map[key]uint64)}
}

// RecoverSequenceId updates the next expected sequence ID for a given worker based on the provided sequence.
func (h *Handler) RecoverSequenceId(seq sequence.Source, clientId string) {
	h.dupsByWorker[newKey(seq.WorkerUuid(), clientId)] = seq.Id() + 1
}

// IsDuplicate determines whether the given sequence ID from a worker is a duplicate.
func (h *Handler) IsDuplicate(seq sequence.Source, clientId string) bool {
	nextSequenceId := h.dupsByWorker[newKey(seq.WorkerUuid(), clientId)]
	if seq.Id() < nextSequenceId {
		return true
	}

	h.RecoverSequenceId(seq, clientId)

	return false
}
