package ioutils

import (
	"bytes"
	"encoding/binary"
	"net"
)

const U8Size = 1
const U64Size = 8

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

func ReadU8FromSlice(buf []byte) (uint8, []byte) {
	return buf[0], buf[U8Size:]
}

func ReadU64FromSlice(buf []byte) (uint64, []byte) {
	return binary.BigEndian.Uint64(buf), buf[U64Size:]
}

func ReadI64(buf *bytes.Buffer) (int64, error) {
	var i int64
	err := binary.Read(buf, binary.BigEndian, &i)
	return i, err
}

func ReadU8(buf *bytes.Buffer) (uint8, error) {
	var i uint8
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
