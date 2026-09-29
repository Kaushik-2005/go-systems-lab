package registry

import (
	"context"
	"sync"
	"time"
)

type Instance struct {
	ID      string
	Address string
}

type registeredInstance struct {
	instance  Instance
	expiresAt time.Time
}

type Registry struct {
	mu       sync.RWMutex
	services map[string]map[string]registeredInstance
}

func New() *Registry {
	return &Registry{
		services: make(map[string]map[string]registeredInstance),
	}
}

func (r *Registry) Register(
	serviceName string,
	instance Instance,
	ttl time.Duration,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.services[serviceName] == nil {
		r.services[serviceName] = make(map[string]registeredInstance)
	}

	r.services[serviceName][instance.ID] = registeredInstance{
		instance:  instance,
		expiresAt: time.Now().Add(ttl),
	}
}

func (r *Registry) Lookup(serviceName string) []Instance {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.removeExpiredLocked()

	registered := r.services[serviceName]
	result := make([]Instance, 0, len(registered))

	for _, entry := range registered {
		result = append(result, entry.instance)
	}

	return result
}

func (r *Registry) Heartbeat(
	serviceName string,
	instanceID string,
	ttl time.Duration,
) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[serviceName]
	entry, found := instances[instanceID]

	if !found || time.Now().After(entry.expiresAt) {
		return false
	}

	entry.expiresAt = time.Now().Add(ttl)
	instances[instanceID] = entry

	return true
}

func (r *Registry) Deregister(serviceName string, instanceID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[serviceName]
	if _, found := instances[instanceID]; !found {
		return false
	}

	delete(instances, instanceID)
	if len(instances) == 0 {
		delete(r.services, serviceName)
	}

	return true
}

func (r *Registry) removeExpiredLocked() {
	now := time.Now()

	for serviceName, instances := range r.services {
		for instanceID, entry := range instances {
			if now.After(entry.expiresAt) {
				delete(instances, instanceID)
			}
		}

		if len(instances) == 0 {
			delete(r.services, serviceName)
		}
	}
}

func (r *Registry) StartCleanup(
	ctx context.Context,
	interval time.Duration,
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.mu.Lock()
				r.removeExpiredLocked()
				r.mu.Unlock()

			case <-ctx.Done():
				return
			}
		}
	}()
}
