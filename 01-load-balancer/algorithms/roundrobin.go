package algorithms

import (
	"sync"

	"load-balancer/servers"
)

// RoundRobin visits each backend in order, then starts again at the first.
type RoundRobin struct {
	mu    sync.Mutex
	index int
}

func (r *RoundRobin) Next(backends []*servers.Server) *servers.Server {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(backends) == 0 {
		return nil
	}

	backend := backends[r.index%len(backends)]
	r.index++
	return backend
}
