package sequence

import (
	"errors"
	"fmt"
	"strconv"

	"tp1/pkg/utils/id"
)

var errNotEnoughArguments = errors.New("not enough arguments")

// Destination represents a tuple of destination key and sequence ID.
type Destination struct {
	key string
	id  uint64
}

// DstNew creates a new Destination from a destination key and a sequence ID.
func DstNew(key string, sequenceId uint64) Destination {
	return Destination{key: key, id: sequenceId}
}

// dstFromString parses a sequence number string into a Destination. If the sequence number's
// format is invalid, an error gets returned.
func dstFromString(seq string) (*Destination, error) {
	parts, err := id.SplitIdAtLastIndex(seq)
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Destination{key: parts[0], id: id}, nil
}

// DstsFromStrings parses many sequence numbers into a slice of Destination. If any sequence number's format is
// invalid, an error gets returned.
func DstsFromStrings(seq []string) ([]Destination, error) {
	if len(seq) == 0 {
		return nil, errNotEnoughArguments
	}

	count, err := strconv.Atoi(seq[0])
	if err != nil {
		return nil, err
	}

	if len(seq) < count+1 {
		return nil, errNotEnoughArguments
	}

	sequenceIds := make([]Destination, 0, count)
	for i := 1; i < count+1; i++ {
		sequenceId, err := dstFromString(seq[i])
		if err != nil {
			return nil, err
		}

		sequenceIds = append(sequenceIds, *sequenceId)
	}

	return sequenceIds, nil
}

// Key returns the Destination underlying destination key.
func (s Destination) Key() string {
	return s.key
}

// Id returns the Destination underlying sequence ID.
func (s Destination) Id() uint64 {
	return s.id
}

// ToString turns the Destination into a human-readable string.
func (s Destination) ToString() string {
	return fmt.Sprintf("%s%s%d", s.key, id.Separator, s.id)
}
