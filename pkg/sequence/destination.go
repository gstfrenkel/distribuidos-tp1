package sequence

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Destination struct {
	key string
	id  uint64
}

func DstNew(key string, sequenceId uint64) Destination {
	return Destination{key: key, id: sequenceId}
}

func DstFromString(seq string) (*Destination, error) {
	lastHyphenIndex := strings.LastIndex(seq, "-")
	if lastHyphenIndex == -1 {
		return nil, errors.New("invalid destination sequence: " + seq)
	}

	key := seq[:lastHyphenIndex]
	id, err := strconv.ParseUint(seq[lastHyphenIndex+1:], 10, 64)
	if err != nil {
		return nil, err
	}

	return &Destination{key: key, id: id}, nil
}

func (s Destination) Key() string {
	return s.key
}

func (s Destination) Id() uint64 {
	return s.id
}

func (s Destination) ToString() string {
	return fmt.Sprintf("%s-%d", s.key, s.id)
}
