package sequence

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Source struct {
	worker uint8
	id     uint64
}

func SrcNew(workerId uint8, sequenceId uint64) Source {
	return Source{worker: workerId, id: sequenceId}
}

func SrcFromString(seq string) (*Source, error) {
	parts := strings.Split(seq, "-")
	if len(parts) != 2 {
		return nil, errors.New("invalid source sequence: " + seq)
	}

	workerId, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return nil, err
	}
	sequenceId, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Source{worker: uint8(workerId), id: sequenceId}, nil
}

func (s Source) WorkerId() uint8 {
	return s.worker
}

func (s Source) Id() uint64 {
	return s.id
}

func (s Source) ToString() string {
	return fmt.Sprintf("%d-%d", s.worker, s.id)
}
