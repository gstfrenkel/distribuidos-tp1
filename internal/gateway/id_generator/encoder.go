package id_generator

import (
	"bytes"
	"encoding/gob"
	"tp1/pkg/logs"
)

var buf bytes.Buffer
var enc = gob.NewEncoder(&buf)

func EncodeClientId(clientId string) []byte {
	buf.Reset()
	err := enc.Encode(clientId)
	if err != nil {
		logs.Logger.Errorf("Error encoding client id: %s", err)
	}
	return buf.Bytes()
}
