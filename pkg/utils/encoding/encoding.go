package encoding

import (
	"bytes"
	"encoding/binary"
)

func EncodeBool(buf *bytes.Buffer, data bool) error {
	return EncodeNumber(buf, data)
}

func DecodeBool(buf *bytes.Buffer) (bool, error) {
	n, err := DecodeUint8(buf)
	return n == 1, err
}

func EncodeNumber(buf *bytes.Buffer, data any) error {
	return binary.Write(buf, binary.BigEndian, data)
}

func DecodeInt8(buf *bytes.Buffer) (int8, error) {
	var data int8
	return data, binary.Read(buf, binary.BigEndian, &data)
}

func DecodeInt64(buf *bytes.Buffer) (int64, error) {
	var data int64
	return data, binary.Read(buf, binary.BigEndian, &data)
}

func DecodeUint8(buf *bytes.Buffer) (uint8, error) {
	var data uint8
	return data, binary.Read(buf, binary.BigEndian, &data)
}

func DecodeUint16(buf *bytes.Buffer) (uint16, error) {
	var data uint16
	return data, binary.Read(buf, binary.BigEndian, &data)
}

func DecodeUint32(buf *bytes.Buffer) (uint32, error) {
	var data uint32
	return data, binary.Read(buf, binary.BigEndian, &data)
}

func DecodeUint64(buf *bytes.Buffer) (uint64, error) {
	var data uint64
	return data, binary.Read(buf, binary.BigEndian, &data)
}

func EncodeString(buf *bytes.Buffer, data string) error {
	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	i := 0
	for i < len(data) {
		n, err := buf.WriteString(data[i:])
		if err != nil {
			return err
		}
		i += n
	}

	return nil
}

// DecodeString decodes a string from a byte slice. The first 32 bits are the len of the string
func DecodeString(buf *bytes.Buffer) (string, error) {
	var size uint32

	if err := binary.Read(buf, binary.BigEndian, &size); err != nil {
		return "", err
	}

	aux := make([]byte, size)
	i := uint32(0)

	for i < size {
		n, err := buf.Read(aux)
		if err != nil {
			return "", err
		}
		i += uint32(n)
	}

	return string(aux), nil
}
