package message

import (
	"bytes"

	"tp1/pkg/utils/encoding"
)

type TextReviews map[int64][]string

func TextReviewFromBytes(b []byte) (TextReviews, error) {
	buf := bytes.NewBuffer(b)

	size, err := encoding.DecodeUint32(buf)
	if err != nil {
		return nil, err
	}

	textReviews := make(TextReviews, size)

	for i := uint32(0); i < size; i++ {
		id, err := encoding.DecodeInt64(buf)
		if err != nil {
			return nil, err
		}

		reviewsCount, err := encoding.DecodeUint32(buf)
		if err != nil {
			return nil, err
		}

		reviews := make([]string, 0, reviewsCount)
		for j := uint32(0); j < reviewsCount; j++ {
			rev, err := encoding.DecodeString(buf)
			if err != nil {
				return nil, err
			}
			reviews = append(reviews, rev)
		}

		textReviews[id] = reviews
	}

	return textReviews, nil
}

func (m TextReviews) ToBytes() ([]byte, error) {
	buf := bytes.Buffer{}

	if err := encoding.EncodeNumber(&buf, uint32(len(m))); err != nil {
		return nil, err
	}

	for id, reviews := range m {
		if err := encoding.EncodeNumber(&buf, id); err != nil {
			return nil, err
		}

		if err := encoding.EncodeNumber(&buf, uint32(len(reviews))); err != nil {
			return nil, err
		}

		for _, rev := range reviews {
			if err := encoding.EncodeString(&buf, rev); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}
