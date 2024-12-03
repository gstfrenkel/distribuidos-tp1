package sequence

import (
	"fmt"
	"strconv"
	"tp1/pkg/utils/encoding"
)

type Source struct {
	workerUuid string
	sequenceId uint64
}

func SrcNew(workerUuid string, sequenceId uint64) Source {
	return Source{workerUuid: workerUuid, sequenceId: sequenceId}
}

func SrcFromString(seq string) (*Source, error) {
	parts, err := encoding.SplitId(seq)
	if err != nil {
		return nil, err
	}

	sequenceId, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Source{workerUuid: parts[0], sequenceId: sequenceId}, nil
}

func (s Source) WorkerUuid() string {
	return s.workerUuid
}

func (s Source) Id() uint64 {
	return s.sequenceId
}

func (s Source) ToString() string {
	return fmt.Sprintf("%s%s%d", s.workerUuid, encoding.Separator, s.sequenceId)
}
