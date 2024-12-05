package message

import (
	"bytes"
	"encoding/gob"
)

const (
	ReviewId Id = iota + 1
	GameId
	EofId
	ScoredReviewId
	ReviewWithTextId
	GameNameId
	GameReleaseId
	PlatformId
	GameWithPlaytimeId
)

type Id uint8

func fromBytes(b []byte, msg any) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(b))
	return decoder.Decode(msg)
}

func toBytes(msg any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
