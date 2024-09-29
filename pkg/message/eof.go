package message

type Eof []uint8

func EofFromBytes(b []byte) (Eof, error) {
	var m Eof
	return m, fromBytes(b, &m)
}

func (m Eof) ToBytes() ([]byte, error) {
	return toBytes(m)
}

func (m Eof) Contains(v uint8) bool {
	for _, w := range m {
		if w == v {
			return true
		}
	}
	return false
}
