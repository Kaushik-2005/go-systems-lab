package cache

import (
	"errors"
	"sync"
	"time"
)

var ErrInvalidCapacity = errors.New("cache capacity must be positive")

type entry struct {
	key       string
	value     string
	expiresAt time.Time
	prev      *entry
	next      *entry
}

type Node struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*entry
	head     *entry
	tail     *entry
}

func NewNode(capacity int) (*Node, error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}
	return &Node{capacity: capacity, items: make(map[string]*entry)}, nil
}

func (n *Node) Set(key, value string, ttl time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if current, ok := n.items[key]; ok {
		current.value = value
		current.expiresAt = expiration(ttl)
		n.moveToFront(current)
		return
	}

	current := &entry{key: key, value: value, expiresAt: expiration(ttl)}
	n.items[key] = current
	n.addToFront(current)

	if len(n.items) > n.capacity {
		n.remove(n.tail)
	}
}

func (n *Node) Get(key string) (string, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	current, ok := n.items[key]
	if !ok {
		return "", false
	}
	if expired(current) {
		n.remove(current)
		return "", false
	}

	n.moveToFront(current)
	return current.value, true
}

func (n *Node) Delete(key string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	current, ok := n.items[key]
	if !ok {
		return false
	}
	n.remove(current)
	return true
}

func (n *Node) Entries() map[string]string {
	n.mu.Lock()
	defer n.mu.Unlock()

	result := make(map[string]string, len(n.items))
	for key, current := range n.items {
		if expired(current) {
			n.remove(current)
			continue
		}
		result[key] = current.value
	}
	return result
}

func (n *Node) Len() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.items)
}

func (n *Node) addToFront(current *entry) {
	current.next = n.head
	if n.head != nil {
		n.head.prev = current
	} else {
		n.tail = current
	}
	n.head = current
}

func (n *Node) moveToFront(current *entry) {
	if current == n.head {
		return
	}
	n.detach(current)
	n.addToFront(current)
}

func (n *Node) detach(current *entry) {
	if current.prev != nil {
		current.prev.next = current.next
	} else {
		n.head = current.next
	}
	if current.next != nil {
		current.next.prev = current.prev
	} else {
		n.tail = current.prev
	}
}

func (n *Node) remove(current *entry) {
	n.detach(current)
	delete(n.items, current.key)
}

func expiration(ttl time.Duration) time.Time {
	if ttl <= 0 {
		return time.Time{}
	}
	return time.Now().Add(ttl)
}

func expired(current *entry) bool {
	return !current.expiresAt.IsZero() && !time.Now().Before(current.expiresAt)
}
