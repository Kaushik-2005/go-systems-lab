package cache

import (
	"errors"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
	"time"
)

var ErrNodeExists = errors.New("cache node already exists")
var ErrNodeNotFound = errors.New("cache node not found")
var ErrNoNodes = errors.New("cluster has no cache nodes")
var ErrInvalidReplication = errors.New("replication factor must be positive")

type Cluster struct {
	mu                sync.RWMutex
	capacity          int
	replicas          int
	replicationFactor int
	nodes             map[string]*namedNode
	ring              []ringPoint
}

type ringPoint struct {
	hash uint32
	node string
}

type namedNode struct {
	name    string
	node    *Node
	healthy bool
}

func NewCluster(names []string, capacity, replicas, replicationFactor int) (*Cluster, error) {
	if capacity <= 0 || replicas <= 0 {
		return nil, ErrInvalidCapacity
	}
	if replicationFactor <= 0 {
		return nil, ErrInvalidReplication
	}

	cluster := &Cluster{
		capacity:          capacity,
		replicas:          replicas,
		replicationFactor: replicationFactor,
		nodes:             make(map[string]*namedNode),
	}
	for _, name := range names {
		if err := cluster.addNodeLocked(name); err != nil {
			return nil, err
		}
	}
	cluster.rebuildRingLocked()
	return cluster, nil
}

func (c *Cluster) AddNode(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.addNodeLocked(name); err != nil {
		return err
	}
	c.rebuildRingLocked()
	return nil
}

func (c *Cluster) RemoveNode(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed, ok := c.nodes[name]
	if !ok {
		return ErrNodeNotFound
	}
	if len(c.nodes) == 1 {
		return ErrNoNodes
	}

	removedEntries := removed.node.Entries()
	delete(c.nodes, name)
	c.rebuildRingLocked()

	for key, value := range removedEntries {
		nodes, err := c.nodesForLocked(key)
		if err != nil {
			return err
		}
		for _, node := range nodes {
			node.node.Set(key, value, 0)
		}
	}
	return nil
}

func (c *Cluster) SetNodeHealth(name string, healthy bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.nodes[name]
	if !ok {
		return ErrNodeNotFound
	}
	node.healthy = healthy
	return nil
}

func (c *Cluster) NodeFor(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodeForLocked(key)
}

func (c *Cluster) ReplicaNodes(key string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes, err := c.nodesForLocked(key)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.name)
	}
	return names, nil
}

func (c *Cluster) Set(key, value string, ttl time.Duration) (string, error) {
	c.mu.RLock()
	nodes, err := c.nodesForLocked(key)
	c.mu.RUnlock()
	if err != nil {
		return "", err
	}

	for _, node := range nodes {
		node.node.Set(key, value, ttl)
	}
	return nodes[0].name, nil
}

func (c *Cluster) Get(key string) (string, bool, string, error) {
	c.mu.RLock()
	nodes, err := c.nodesForLocked(key)
	c.mu.RUnlock()
	if err != nil {
		return "", false, "", err
	}

	value, found := nodes[0].node.Get(key)
	return value, found, nodes[0].name, nil
}

func (c *Cluster) Rebalance() (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.ring) == 0 {
		return 0, ErrNoNodes
	}

	names := make([]string, 0, len(c.nodes))
	for name := range c.nodes {
		names = append(names, name)
	}
	sort.Strings(names)

	values := make(map[string]string)
	oldOwners := make(map[string]string)
	for _, name := range names {
		for key, value := range c.nodes[name].node.Entries() {
			if _, exists := values[key]; !exists {
				values[key] = value
				oldOwners[key] = name
			}
		}
	}

	moved := 0
	for key, value := range values {
		nodes, err := c.nodesForLocked(key)
		if err != nil {
			return 0, err
		}
		desired := make(map[string]bool, len(nodes))
		for _, node := range nodes {
			desired[node.name] = true
			node.node.Set(key, value, 0)
		}
		if oldOwners[key] != nodes[0].name {
			moved++
		}
		for _, name := range names {
			if !desired[name] {
				c.nodes[name].node.Delete(key)
			}
		}
	}
	return moved, nil
}

func (c *Cluster) NodeNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.nodes))
	for name := range c.nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Cluster) addNodeLocked(name string) error {
	if name == "" {
		return ErrNodeNotFound
	}
	if _, ok := c.nodes[name]; ok {
		return ErrNodeExists
	}

	node, err := NewNode(c.capacity)
	if err != nil {
		return err
	}
	c.nodes[name] = &namedNode{name: name, node: node, healthy: true}
	return nil
}

func (c *Cluster) rebuildRingLocked() {
	c.ring = nil
	names := make([]string, 0, len(c.nodes))
	for name := range c.nodes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		for replica := 0; replica < c.replicas; replica++ {
			c.ring = append(c.ring, ringPoint{
				hash: hashString(name + "#" + strconv.Itoa(replica)),
				node: name,
			})
		}
	}
	sort.Slice(c.ring, func(i, j int) bool {
		return c.ring[i].hash < c.ring[j].hash
	})
}

func (c *Cluster) nodeForLocked(key string) (string, error) {
	nodes, err := c.nodesForLocked(key)
	if err != nil {
		return "", err
	}
	return nodes[0].name, nil
}

func (c *Cluster) nodesForLocked(key string) ([]*namedNode, error) {
	if len(c.ring) == 0 {
		return nil, ErrNoNodes
	}

	hash := hashString(key)
	index := sort.Search(len(c.ring), func(index int) bool {
		return c.ring[index].hash >= hash
	})
	if index == len(c.ring) {
		index = 0
	}

	limit := c.replicationFactor
	if limit > len(c.nodes) {
		limit = len(c.nodes)
	}
	result := make([]*namedNode, 0, limit)
	seen := make(map[string]bool, limit)
	for offset := 0; len(result) < limit && offset < len(c.ring); offset++ {
		point := c.ring[(index+offset)%len(c.ring)]
		if seen[point.node] || !c.nodes[point.node].healthy {
			continue
		}
		seen[point.node] = true
		result = append(result, c.nodes[point.node])
	}
	if len(result) == 0 {
		return nil, ErrNoNodes
	}
	return result, nil
}

func hashString(value string) uint32 {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(value))
	return hash.Sum32()
}
