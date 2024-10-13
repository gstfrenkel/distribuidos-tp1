package message

import "fmt"

type Platform struct {
	Windows int
	Linux   int
	Mac     int
}

func (p Platform) ToBytes() ([]byte, error) {
	return toBytes(p)
}

func (p *Platform) Increment(other Platform) {
	p.Windows += other.Windows
	p.Linux += other.Linux
	p.Mac += other.Mac
}

func PlatfromFromBytes(b []byte) (Platform, error) {
	var m Platform
	return m, fromBytes(b, &m)
}

func (p *Platform) ResetValues() {
	p.Linux = 0
	p.Windows = 0
	p.Mac = 0
}

func (p Platform) IsEmpty() bool {
	return p.Windows == 0 && p.Linux == 0 && p.Mac == 0
}

func (p Platform) ToResultString() string {
	return fmt.Sprintf("Q1: Juegos soportados en cada plataforma\nWindows: [%d], Linux: [%d], Mac: [%d]", p.Windows, p.Linux, p.Mac)
}
