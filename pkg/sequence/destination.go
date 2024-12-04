package sequence

import (
	"errors"
	"fmt"
	"strconv"

	"tp1/pkg/utils/id_generator"
)

var errNotEnoughArguments = errors.New("not enough arguments")

type Destination struct {
	key string
	id  uint64
}

func DstNew(key string, sequenceId uint64) Destination {
	return Destination{key: key, id: sequenceId}
}

func dstFromString(seq string) (*Destination, error) {
	parts, err := id_generator.SplitIdAtLastIndex(seq)
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Destination{key: parts[0], id: id}, nil
}

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
	for i := 1; i < count; i++ {
		sequenceId, err := dstFromString(seq[i])
		if err != nil {
			return nil, err
		}

		sequenceIds = append(sequenceIds, *sequenceId)
	}

	return sequenceIds, nil
}

func (s Destination) Key() string {
	return s.key
}

func (s Destination) Id() uint64 {
	return s.id
}

func (s Destination) ToString() string {
	return fmt.Sprintf("%s%s%d", s.key, id_generator.Separator, s.id)
}
