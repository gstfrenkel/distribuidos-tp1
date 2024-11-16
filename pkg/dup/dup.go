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
	h.dupsByWorker[seq.WorkerId()] = seq.Id()
}

func (h *handler) IsDuplicate(seq sequence.Source) bool {
	nextSequenceId := h.dupsByWorker[seq.WorkerId()]
	if seq.Id() < nextSequenceId {
		return true
	} /*else if seq.Id() > nextSequenceId {
		logs.Logger.Errorf("skipped sequence number")
	}
	Está mal porque puede pasar que un productor meta cosas en una cola de la cual consumen 2 filtros, y el núm de seq
	1 puede ir a parar al primer filtro, el núm de seq 2 al segundo, y el núm de seq 3 al primero de vuelta.
	*/

	h.dupsByWorker[seq.WorkerId()] = seq.Id() + 1

	return false
}
