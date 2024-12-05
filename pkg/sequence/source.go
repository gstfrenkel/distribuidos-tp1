package sequence

import (
	"fmt"
	"strconv"

	"tp1/pkg/utils/id"
)

// Source represents a tuple of worker UUID and sequence ID.
type Source struct {
	workerUuid string
	sequenceId uint64
}

// SrcNew creates a new Source from a worker's UUID and a sequence ID.
func SrcNew(workerUuid string, sequenceId uint64) Source {
	return Source{workerUuid: workerUuid, sequenceId: sequenceId}
}

// SrcFromString parses a sequence number string into a Destination. If the sequence number's
// format is invalid, an error gets returned.
func SrcFromString(seq string) (*Source, error) {
	parts, err := id.SplitIdAtLastIndex(seq)
	if err != nil {
		return nil, err
	}

	sequenceId, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Source{workerUuid: parts[0], sequenceId: sequenceId}, nil
}

// WorkerUuid returns the Source underlying worker UUID.
func (s Source) WorkerUuid() string {
	return s.workerUuid
}

// Id returns the Source underlying sequence ID.
func (s Source) Id() uint64 {
	return s.sequenceId
}

// ToString turns the Source into a human-readable string.
func (s Source) ToString() string {
	return fmt.Sprintf("%s%s%d", s.workerUuid, id.Separator, s.sequenceId)
}
