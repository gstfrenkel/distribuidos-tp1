package ioutils

import (
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

func ReadU8FromSlice(buf []byte) uint8 {
	return buf[0]
}

func ReadU64FromSlice(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}
