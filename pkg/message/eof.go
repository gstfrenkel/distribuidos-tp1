package message

import (
	"bytes"
	"encoding/binary"
)

type Eof []uint8

func (m Eof) ToBytes() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, uint8(len(m))); err != nil {
		return nil, err
	}

	for _, w := range m {
		if err := binary.Write(&buf, binary.BigEndian, w); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func EofFromBytes(b []byte) (Eof, error) {
	buf := bytes.NewBuffer(b)
	var size uint8
	if err := binary.Read(buf, binary.BigEndian, &size); err != nil {
		return nil, err
	}

	m := make(Eof, size)
	for i := 0; i < int(size); i++ {
		if err := binary.Read(buf, binary.BigEndian, &m[i]); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m Eof) Contains(v uint8) bool {
	for _, w := range m {
		if w == v {
			return true
		}
	}
	return false
}
