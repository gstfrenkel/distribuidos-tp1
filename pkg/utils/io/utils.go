package io

import (
	"io"
	"net"
	"os/exec"
	"strings"
	"tp1/pkg/logs"
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

// MoveBuff moves the buffer n positions to the left keeping the original capacity
func MoveBuff(data []byte, n int) []byte {
	copy(data, data[n:])
	return data[:len(data)-n]
}

// ExecCommand executes a command in the shell
func ExecCommand(command string) error {
	logs.Logger.Infof("Executing command: %s", command)
	commands := strings.Split(command, " ")
	return exec.Command(commands[0], commands[1:]...).Run()
}
