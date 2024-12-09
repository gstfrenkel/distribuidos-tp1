package test_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"tp1/pkg/message"
)

func Test_TextReviews(t *testing.T) {
	exp := message.TextReviews{
		12:   []string{"Test one", "Twelve", "Test one"},
		1254: []string{"One", "Two", "Five", "Four"},
		1134: []string{""},
		1:    []string{},
	}

	b, err := exp.ToBytes()
	assert.NoError(t, err)

	res, err := message.TextReviewFromBytes(b)
	assert.NoError(t, err)

	assert.Equal(t, exp, res)
}
