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

// ExecCommand executes a command in the shell and returns its output
func ExecCommand(command string) (string, error) {
	logs.Logger.Infof("Executing command: %s", command)
	commands := strings.Split(command, " ")
	cmd := exec.Command(commands[0], commands[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
