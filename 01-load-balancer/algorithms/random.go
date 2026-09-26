package algorithms

import (
	"math/rand"
	"sync"

	"load-balancer/servers"
)

// Random selects one backend uniformly from the supplied list.
type Random struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func NewRandom() *Random {
	return &Random{rng: rand.New(rand.NewSource(rand.Int63()))}
}

func (r *Random) Next(backends []*servers.Server) *servers.Server {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(backends) == 0 {
		return nil
	}

	return backends[r.rng.Intn(len(backends))]
}
