package encoding

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DecodeInt8(t *testing.T) {
	buf := bytes.Buffer{}
	exp := int8(10)

	err := EncodeNumber(&buf, exp)
	assert.NoError(t, err)

	res, err := DecodeInt8(&buf)
	assert.NoError(t, err)
	assert.Equal(t, exp, res)
}

func Test_DecodeUint8(t *testing.T) {
	buf := bytes.Buffer{}
	exp := uint8(10)

	err := EncodeNumber(&buf, exp)
	assert.NoError(t, err)

	res, err := DecodeUint8(&buf)
	assert.NoError(t, err)
	assert.Equal(t, exp, res)
}

func Test_DecodeUint16(t *testing.T) {
	buf := bytes.Buffer{}
	exp := uint16(105)

	err := EncodeNumber(&buf, exp)
	assert.NoError(t, err)

	res, err := DecodeUint16(&buf)
	assert.NoError(t, err)
	assert.Equal(t, exp, res)
}

func Test_DecodeUint32(t *testing.T) {
	buf := bytes.Buffer{}
	exp := uint32(101343)

	err := EncodeNumber(&buf, exp)
	assert.NoError(t, err)

	res, err := DecodeUint32(&buf)
	assert.NoError(t, err)
	assert.Equal(t, exp, res)
}

func Test_DecodeUint64(t *testing.T) {
	buf := bytes.Buffer{}
	exp := uint64(104358934095)

	err := EncodeNumber(&buf, exp)
	assert.NoError(t, err)

	res, err := DecodeUint64(&buf)
	assert.NoError(t, err)
	assert.Equal(t, exp, res)
}

func Test_String(t *testing.T) {
	buf := bytes.Buffer{}
	exp := "encoding/decoding test string"

	err := EncodeString(&buf, exp)
	assert.Nil(t, err)

	res, err := DecodeString(&buf)
	assert.Nil(t, err)
	assert.Equal(t, exp, res)
}

func Test_Bool(t *testing.T) {
	buf1 := bytes.Buffer{}
	buf2 := bytes.Buffer{}
	exp1 := true
	exp2 := false

	err := EncodeBool(&buf1, exp1)
	assert.Nil(t, err)
	err = EncodeBool(&buf2, exp2)
	assert.Nil(t, err)

	res1, err := DecodeBool(&buf1)
	assert.Nil(t, err)
	res2, err := DecodeBool(&buf2)
	assert.Nil(t, err)
	assert.Equal(t, exp1, res1)
	assert.Equal(t, exp2, res2)
}
