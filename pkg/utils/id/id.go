package id

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"tp1/pkg/logs"
	ioutils "tp1/pkg/utils/io"
)

func errInvalidId(id string) error {
	return fmt.Errorf("invalid id: %s", id)
}

const (
	ClientIdLen = 32
	defaultFile = "id-generator-%d.csv"
	Separator   = "-"
	idParts     = 2
)

type Generator struct {
	file   *ioutils.File
	prefix uint8
	nextId uint16
}

func NewGenerator(prefix uint8, fileName string) *Generator {
	if fileName == "" {
		fileName = defaultFile
	}

	file, err := ioutils.NewFile(fmt.Sprintf(fileName, prefix))
	if err != nil {
		logs.Logger.Errorf("Error creating file: %v", err)
		return nil
	}

	return &Generator{
		file:   file,
		prefix: prefix,
		nextId: loadFromDisk(file),
	}
}

// GetId returns a new id. The format of the id is prefix-nextId
func (g *Generator) GetId() string {
	id := strconv.Itoa(int(g.prefix)) + Separator + strconv.Itoa(int(g.nextId))
	g.nextId++

	err := g.writeToDisk()
	if err != nil {
		logs.Logger.Errorf("Error writing id to file: %v", err)
	}

	return id
}

func (g *Generator) writeToDisk() error {
	return g.file.Overwrite([]string{strconv.Itoa(int(g.nextId))})
}

func (g *Generator) Close() {
	g.file.Close()
}

func loadFromDisk(file *ioutils.File) uint16 {
	nextId := uint16(0)

	line, err := file.Read()
	if err != nil {
		return nextId
	}

	id, err := strconv.ParseUint(line[0], 10, 16)
	if err != nil {
		logs.Logger.Errorf("Error parsing id: %v", err)
	} else {
		nextId = uint16(id)
	}

	return nextId
}

// EncodeClientId encodes a client id to a fixed-length byte slice of 32 bytes
func EncodeClientId(clientId string) []byte {
	clientIdBytes := []byte(clientId)
	clientIdBytesLen := len(clientIdBytes)

	if clientIdBytesLen > ClientIdLen {
		logs.Logger.Errorf("Error id clientId: clientId too long")
		return nil
	}

	buff := make([]byte, ClientIdLen)
	copy(buff, clientIdBytes)

	return buff
}

func DecodeClientId(data []byte) string {
	return string(bytes.TrimRight(data, "\x00"))
}

// SplitId splits a string into 2 parts by Separator.
func SplitId(id string) ([idParts]string, error) {
	parts := strings.Split(id, Separator)
	if len(parts) != idParts {
		return [idParts]string{}, errInvalidId(id)
	}

	return [idParts]string{parts[0], parts[1]}, nil
}

// SplitIdAtLastIndex splits a string into 2 parts by the last Separator.
func SplitIdAtLastIndex(id string) ([idParts]string, error) {
	index := strings.LastIndex(id, Separator)
	if index == -1 {
		return [idParts]string{}, errInvalidId(id)
	}

	return [idParts]string{id[:index], id[index+1:]}, nil
}
