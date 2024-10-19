package id_generator

import (
	"bytes"
	"tp1/pkg/logs"
)

const ClientIdLen = 32

// EncodeClientId encodes a client id to a fixed-length byte slice of 32 bytes
func EncodeClientId(clientId string) []byte {
	clientIdBytes := []byte(clientId)
	clientIdBytesLen := len(clientIdBytes)

	if clientIdBytesLen > ClientIdLen {
		logs.Logger.Errorf("Error encoding clientId: clientId too long")
		return nil
	}

	buff := make([]byte, ClientIdLen)
	copy(buff, clientIdBytes)

	return buff
}

func DecodeClientId(data []byte) string {
	return string(bytes.TrimRight(data, "\x00"))
}
