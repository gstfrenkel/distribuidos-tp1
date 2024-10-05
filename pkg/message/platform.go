package message

type Platform struct {
	Windows int
	Linux   int
	Mac     int
}

func (p Platform) ToBytes() ([]byte, error) {
	return toBytes(p)
}
