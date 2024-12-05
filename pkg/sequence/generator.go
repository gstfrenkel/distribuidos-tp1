package sequence

type Generator struct {
	sequenceIds map[string]uint64
}

// NewGenerator creates a new sequence ID generator.
func NewGenerator() *Generator {
	return &Generator{sequenceIds: make(map[string]uint64)}
}

// RecoverId recovers a lost sequence ID.
func (g *Generator) RecoverId(destination Destination) {
	g.sequenceIds[destination.Key()] = destination.Id() + 1
}

// NextId generates and returns the next sequence ID for a given key.
func (g *Generator) NextId(key string) uint64 {
	sequenceId, ok := g.sequenceIds[key]
	if !ok {
		g.sequenceIds[key] = 1
	} else {
		g.sequenceIds[key]++
	}
	return sequenceId
}
