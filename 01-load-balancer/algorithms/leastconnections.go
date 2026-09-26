package algorithms

import (
	"sync"

	"load-balancer/servers"
)

// LeastConnections selects the backend with the fewest active requests.
type LeastConnections struct {
	mu sync.Mutex
}

func (l *LeastConnections) Next(backends []*servers.Server) *servers.Server {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(backends) == 0 {
		return nil
	}

	selected := backends[0]
	for _, backend := range backends[1:] {
		if backend.ActiveRequests() < selected.ActiveRequests() {
			selected = backend
		}
	}
	return selected
}
