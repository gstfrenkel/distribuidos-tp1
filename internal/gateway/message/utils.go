package message

import (
	"bytes"
	"encoding/binary"
)

func WriteBytesToBuff(fields []interface{}, buf *bytes.Buffer) error {
	for _, field := range fields {
		if err := binary.Write(buf, binary.BigEndian, field); err != nil {
			return err
		}
	}
	return nil
}

// ReadBytesFromBuff Reads fixed size fields from a buffer
func ReadBytesFromBuff(fields []interface{}, buf *bytes.Buffer) error {
	for _, field := range fields {
		if err := binary.Read(buf, binary.BigEndian, field); err != nil {
			return err
		}
	}
	return nil
}

func ReadI64(buf *bytes.Buffer) (int64, error) {
	var i int64
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func ReadU64(buf *bytes.Buffer) (uint64, error) {
	var i uint64
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func ReadString(buf *bytes.Buffer, length uint64) (string, error) {
	b := make([]byte, length)
	if err := binary.Read(buf, binary.BigEndian, b); err != nil {
		return "", err
	}
	return string(b), nil
}
