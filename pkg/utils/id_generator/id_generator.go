package id_generator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"tp1/pkg/logs"
	utilsio "tp1/pkg/utils/io"
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

type IdGenerator struct {
	file   *utilsio.File
	prefix uint8
	nextId uint16
}

func New(prefix uint8, fileName string) *IdGenerator {
	if fileName == "" {
		fileName = defaultFile
	}

	file, err := utilsio.NewFile(fmt.Sprintf(fileName, prefix))
	if err != nil {
		logs.Logger.Errorf("Error creating file: %v", err)
		return nil
	}

	return &IdGenerator{
		file:   file,
		prefix: prefix,
		nextId: loadFromDisk(file),
	}
}

func loadFromDisk(file *utilsio.File) uint16 {
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

// GetId returns a new id. The format of the id is prefix-nextId
func (g *IdGenerator) GetId() string {
	id := strconv.Itoa(int(g.prefix)) + Separator + strconv.Itoa(int(g.nextId))
	g.nextId++

	err := g.writeToDisk()
	if err != nil {
		logs.Logger.Errorf("Error writing id to file: %v", err)
	}

	return id
}

func (g *IdGenerator) writeToDisk() error {
	return g.file.Overwrite([]string{strconv.Itoa(int(g.nextId))})
}

// EncodeClientId encodes a client id to a fixed-length byte slice of 32 bytes
func EncodeClientId(clientId string) []byte {
	clientIdBytes := []byte(clientId)
	clientIdBytesLen := len(clientIdBytes)

	if clientIdBytesLen > ClientIdLen {
		logs.Logger.Errorf("Error id_generator clientId: clientId too long")
		return nil
	}

	buff := make([]byte, ClientIdLen)
	copy(buff, clientIdBytes)

	return buff
}

func DecodeClientId(data []byte) string {
	return string(bytes.TrimRight(data, "\x00"))
}

func SplitId(id string) ([idParts]string, error) {
	parts := strings.Split(id, Separator)
	if len(parts) != idParts {
		return [idParts]string{}, errInvalidId(id)
	}

	return [idParts]string{parts[0], parts[1]}, nil
}

func SplitIdAtLastIndex(id string) ([idParts]string, error) {
	index := strings.LastIndex(id, Separator)
	if index == -1 {
		return [idParts]string{}, errInvalidId(id)
	}

	return [idParts]string{id[:index], id[index+1:]}, nil
}

func (g *IdGenerator) Close() {
	g.file.Close()
}
