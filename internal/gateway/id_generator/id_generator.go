package id_generator

import "strconv"

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
