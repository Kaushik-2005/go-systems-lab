package ring

import (
	"hash/fnv"
)

// Modulo routes a key using hash(key) modulo the number of nodes.
type Modulo struct {
	nodes []string
}

func NewModulo(nodes []string) *Modulo {
	return &Modulo{nodes: append([]string(nil), nodes...)}
}

func (m *Modulo) Lookup(key string) string {
	if len(m.nodes) == 0 {
		return ""
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	index := hasher.Sum32() % uint32(len(m.nodes))
	return m.nodes[index]
}
