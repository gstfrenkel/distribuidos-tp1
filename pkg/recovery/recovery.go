package recovery

import (
	"fmt"
	"strconv"
	"strings"
	"tp1/pkg/ioutils"

	"tp1/pkg/logs"
)

type Recovery interface {
	IsDuplicate(seq string) bool
	Close()
}

type recovery struct {
	file         *ioutils.File
	dupsByWorker map[uint]uint64
}

func New() (Recovery, error) {
	file, err := ioutils.NewFile()
	if err != nil {
		return nil, err
	}

	return &recovery{
		file:         file,
		dupsByWorker: make(map[uint]uint64),
	}, nil
}

func (r *recovery) RecoverNextLine(value any, parse func([]string, any) error) error {
	line, err := r.file.ReadNext()
	if err != nil {
		return err
	}

	return parse(line, value)
}

func (r *recovery) Log() {

}

func (r *recovery) IsDuplicate(seq string) bool {
	workerId, seqnum := explodeSeq(seq)
	nextSeqnum, ok := r.dupsByWorker[workerId]
	if !ok {
		if seqnum > 1 {
			logs.Logger.Errorf("skipped first sequence number")
		}
		r.dupsByWorker[workerId] = seqnum + 1
		return false
	}

	if seqnum < nextSeqnum {
		return true
	} else if seqnum > nextSeqnum {
		logs.Logger.Errorf("skipped sequence number")
	}
	r.dupsByWorker[workerId] = nextSeqnum + 1
	return false
}

func (r *recovery) Close() {
	r.file.Close()
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
