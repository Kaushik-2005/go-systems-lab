package shard

import (
	"errors"
	"hash/fnv"
	"sync"
)

var ErrNoShards = errors.New("router has no shards")
var ErrShardExists = errors.New("shard already exists")
var ErrShardNotFound = errors.New("shard not found")

type Router struct {
	mu     sync.RWMutex
	shards []string
}

func New(shards []string) *Router {
	shardCopy := append([]string(nil), shards...)
	return &Router{shards: shardCopy}
}

func (r *Router) Route(key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.shards) == 0 {
		return "", ErrNoShards
	}

	index := int(hashKey(key) % uint32(len(r.shards)))
	return r.shards[index], nil
}

func (r *Router) AddShard(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shard := range r.shards {
		if shard == name {
			return ErrShardExists
		}
	}
	r.shards = append(r.shards, name)
	return nil
}

func (r *Router) RemoveShard(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, shard := range r.shards {
		if shard != name {
			continue
		}
		r.shards = append(r.shards[:index], r.shards[index+1:]...)
		return nil
	}
	return ErrShardNotFound
}

func (r *Router) Shards() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return append([]string(nil), r.shards...)
}

func hashKey(key string) uint32 {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	return hash.Sum32()
}
