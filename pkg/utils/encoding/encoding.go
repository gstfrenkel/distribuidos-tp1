package encoding

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"tp1/pkg/logs"
)

func errInvalidId(id string) error {
	return fmt.Errorf("invalid id: %s", id)
}

const (
	ClientIdLen = 32

	Separator = "-"
)

type IdGenerator struct {
	prefix uint8
	nextId uint16
}

func New(prefix uint8) *IdGenerator {
	return &IdGenerator{
		prefix: prefix,
		nextId: 0,
	}
}

// GetId returns a new id. The format of the id is prefix-nextId
func (g *IdGenerator) GetId() string {
	id := strconv.Itoa(int(g.prefix)) + "-" + strconv.Itoa(int(g.nextId))
	g.nextId++

	return id
}

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

func SplitId(id string) ([2]string, error) {
	parts := strings.Split(id, Separator)
	if len(parts) != 2 {
		return [2]string{}, errInvalidId(id)
	}

	return [2]string{parts[0], parts[1]}, nil
}
