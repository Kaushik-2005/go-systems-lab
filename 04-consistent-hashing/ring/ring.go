package ring

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sort"
)

// HashRing places nodes and keys on the same circular hash space.
type HashRing struct {
	points   []uint32
	owners   map[uint32]string
	nodes    map[string][]uint32
	replicas int
}

func New(replicaCounts ...int) *HashRing {
	replicas := 1
	if len(replicaCounts) > 0 && replicaCounts[0] > 0 {
		replicas = replicaCounts[0]
	}

	return &HashRing{
		owners:   make(map[uint32]string),
		nodes:    make(map[string][]uint32),
		replicas: replicas,
	}
}

func (r *HashRing) AddNode(node string) {
	if _, exists := r.nodes[node]; exists {
		return
	}

	points := make([]uint32, 0, r.replicas)
	for replica := 0; replica < r.replicas; replica++ {
		point := hash(fmt.Sprintf("%s#%d", node, replica))
		for {
			if _, exists := r.owners[point]; !exists {
				break
			}
			point++
		}
		points = append(points, point)
		r.owners[point] = node
		r.points = append(r.points, point)
	}
	r.nodes[node] = points
	sort.Slice(r.points, func(i, j int) bool {
		return r.points[i] < r.points[j]
	})
}

func (r *HashRing) RemoveNode(node string) {
	points, exists := r.nodes[node]
	if !exists {
		return
	}

	delete(r.nodes, node)
	for _, point := range points {
		delete(r.owners, point)
		index := sort.Search(len(r.points), func(i int) bool {
			return r.points[i] >= point
		})
		r.points = append(r.points[:index], r.points[index+1:]...)
	}
}

func (r *HashRing) Lookup(key string) string {
	if len(r.points) == 0 {
		return ""
	}

	point := hash(key)
	index := sort.Search(len(r.points), func(i int) bool {
		return r.points[i] >= point
	})
	if index == len(r.points) {
		index = 0
	}
	return r.owners[r.points[index]]
}

func hash(value string) uint32 {
	digest := sha1.Sum([]byte(value))
	return binary.BigEndian.Uint32(digest[:4])
}
