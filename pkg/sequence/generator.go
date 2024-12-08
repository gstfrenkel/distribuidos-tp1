package sequence

type key struct {
	outputKey string
	clientId  string
}

func newKey(outputKey string, clientId string) key {
	return key{outputKey: outputKey, clientId: clientId}
}

type Generator struct {
	sequenceIds map[key]uint64
}

// NewGenerator creates a new sequence ID generator.
func NewGenerator() *Generator {
	return &Generator{sequenceIds: make(map[key]uint64)}
}

// RecoverId recovers a lost sequence ID.
func (g *Generator) RecoverId(destination Destination, clientId string) {
	g.sequenceIds[newKey(destination.Key(), clientId)] = destination.Id() + 1
}

// NextId generates and returns the next sequence ID for a given key.
func (g *Generator) NextId(key string, clientId string) uint64 {
	k := newKey(key, clientId)

	sequenceId, ok := g.sequenceIds[k]
	if !ok {
		g.sequenceIds[k] = 1
	} else {
		g.sequenceIds[k]++
	}
	return sequenceId
}
