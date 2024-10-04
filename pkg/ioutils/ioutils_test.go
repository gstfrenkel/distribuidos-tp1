package ioutils

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadU8FromSlice_ValidInput(t *testing.T) {
	buf := []byte{0x01, 0x02, 0x03}
	val, remaining := ReadU8FromSlice(buf)
	assert.Equal(t, uint8(0x01), val)
	assert.Equal(t, []byte{0x02, 0x03}, remaining)
}

func TestReadU64FromSlice_ValidInput(t *testing.T) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, 0x0102030405060708)
	val, remaining := ReadU64FromSlice(buf)
	assert.Equal(t, uint64(0x0102030405060708), val)
	assert.Equal(t, []byte{}, remaining)
}
