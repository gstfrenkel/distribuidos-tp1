package ioutils

import (
	"encoding/binary"
	"io"
	"net"
	"os/exec"
)

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

func ReadFull(conn net.Conn, buffer []byte, n int) error {
	totalBytesRead := 0

	for totalBytesRead < n {
		bytesRead, err := conn.Read(buffer[totalBytesRead:])
		if err != nil {
			return err
		}
		if bytesRead == 0 {
			return io.EOF
		}
		totalBytesRead += bytesRead
	}

	return nil
}

func ReadU32FromSlice(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

// MoveBuff moves the buffer n positions to the left keeping the original capacity
func MoveBuff(data []byte, n int) []byte {
	copy(data, data[n:])
	return data[:len(data)-n]
}

// ExecCommand executes a command in the shell
func ExecCommand(command []string) error {
	return exec.Command(command[0], command[1:]...).Run()
}
