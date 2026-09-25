package algorithms

import "load-balancer/servers"

// Selector chooses one backend from the configured backends.
type Selector interface {
	Next([]*servers.Server) *servers.Server
}
