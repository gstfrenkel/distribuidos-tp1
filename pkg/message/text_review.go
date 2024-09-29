package message

import (
	"tp1/pkg/message/utils"
)

type TextReview map[utils.Key][]string

func TextReviewFromBytes(b []byte) (TextReview, error) {
	var m TextReview
	return m, fromBytes(b, &m)
}

func (m TextReview) ToBytes() ([]byte, error) {
	return toBytes(m)
}
