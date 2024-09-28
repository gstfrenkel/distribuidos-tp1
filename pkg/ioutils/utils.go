package ioutils

import (
	"bytes"
	"encoding/binary"
	"net"
)

const U8Size = 1
const U64Size = 8
const I64Size = 8

func WriteBytesToBuff(fields []interface{}, buf *bytes.Buffer) error {
	for _, field := range fields {
		if err := binary.Write(buf, binary.BigEndian, field); err != nil {
			return err
		}
	}
	return nil
}

// SendAll Sends all data to a connection socket
func SendAll(conn net.Conn, data []byte) error {
	total := len(data)
	for total > 0 {
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
		total -= n
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

func ReadU8FromSlice(buf []byte) uint8 {
	msg := buf[0]
	buf = buf[U8Size:]
	return msg
}

func ReadU64FromSlice(buf []byte) uint64 {
	u64 := binary.BigEndian.Uint64(buf)
	buf = buf[U64Size:]
	return u64
}

func ReadI64FromSlice(c []byte) int64 {
	i, _ := ReadI64(bytes.NewBuffer(c))
	return i
}

func ReadStringFromSlice(buf []byte, length uint64) string {
	str := string(buf[:length])
	buf = buf[length:]
	return str
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
