package shard

import "sync"

type ShardMap struct {
	mu     sync.RWMutex
	router *Router
	values map[string]map[string]string
}

func NewMap(router *Router) *ShardMap {
	values := make(map[string]map[string]string)
	for _, name := range router.Shards() {
		values[name] = make(map[string]string)
	}
	return &ShardMap{router: router, values: values}
}

func (m *ShardMap) Put(key, value string) error {
	shard, err := m.router.Route(key)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.values[shard]; !ok {
		m.values[shard] = make(map[string]string)
	}
	m.values[shard][key] = value
	return nil
}

func (m *ShardMap) Get(key string) (string, bool, string, error) {
	shard, err := m.router.Route(key)
	if err != nil {
		return "", false, "", err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	value, found := m.values[shard][key]
	return value, found, shard, nil
}

func (m *ShardMap) Rebalance() (int, error) {
	shards := m.router.Shards()
	if len(shards) == 0 {
		return 0, ErrNoShards
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	rebuilt := make(map[string]map[string]string, len(shards))
	for _, name := range shards {
		rebuilt[name] = make(map[string]string)
	}

	moved := 0
	for oldShard, values := range m.values {
		for key, value := range values {
			newShard, err := m.router.Route(key)
			if err != nil {
				return 0, err
			}
			if newShard != oldShard {
				moved++
			}
			rebuilt[newShard][key] = value
		}
	}

	m.values = rebuilt
	return moved, nil
}
