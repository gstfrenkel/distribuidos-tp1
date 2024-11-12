package dup

import (
	"fmt"
	"strconv"
	"strings"
	"tp1/pkg/logs"
)

type Handler interface {
	Add(workerId uint, seqnum uint64)
	IsDuplicate(seq string) bool
}

type handler struct {
	dupsByWorker map[uint]uint64
}

func NewHandler() Handler {
	return &handler{dupsByWorker: make(map[uint]uint64)}
}

func (h *handler) Add(workerId uint, seqnum uint64) {
	h.dupsByWorker[workerId] = seqnum
}

func (h *handler) IsDuplicate(seq string) bool {
	workerId, seqnum := explodeSeq(seq)
	nextSeqnum, ok := h.dupsByWorker[workerId]
	if !ok {
		if seqnum > 1 {
			logs.Logger.Errorf("skipped first sequence number")
		}
		h.dupsByWorker[workerId] = seqnum + 1
		return false
	}

	if seqnum < nextSeqnum {
		return true
	} else if seqnum > nextSeqnum {
		logs.Logger.Errorf("skipped sequence number")
	}
	h.dupsByWorker[workerId] = nextSeqnum + 1
	return false
}

func explodeSeq(seq string) (uint, uint64) {
	parts := strings.Split(seq, "-")
	if len(parts) != 2 {
		return 0, 0
	}

	workerId, err := strconv.ParseUint(parts[0], 10, 0)
	if err != nil {
		logs.Logger.Errorf(fmt.Sprintf("failed to parse worker ID: %s", err.Error()))
		return 0, 0
	}
	seqnum, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		logs.Logger.Errorf(fmt.Sprintf("failed to parse sequence number: %s", err.Error()))
		return 0, 0
	}
	return uint(workerId), seqnum
}
