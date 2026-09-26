package cache

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	key       string
	value     string
	expiresAt time.Time
	prev      *entry
	next      *entry
}

// Cache stores a fixed number of values using least-recently-used eviction.
type Cache struct {
	mu       sync.Mutex
	capacity int
	values   map[string]*entry
	head     *entry
	tail     *entry
}

func New(capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		values:   make(map[string]*entry),
	}
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value, time.Time{})
}

func (c *Cache) SetWithTTL(key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value, time.Now().Add(ttl))
}

func (c *Cache) set(key, value string, expiresAt time.Time) {

	if c.capacity <= 0 {
		return
	}

	if item, ok := c.values[key]; ok {
		item.value = value
		item.expiresAt = expiresAt
		c.moveToFront(item)
		return
	}

	item := &entry{key: key, value: value, expiresAt: expiresAt}
	c.values[key] = item
	c.addToFront(item)

	if len(c.values) > c.capacity {
		oldest := c.tail
		c.remove(oldest)
		delete(c.values, oldest.key)
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.values[key]
	if !ok {
		return "", false
	}
	if c.expired(item) {
		c.remove(item)
		delete(c.values, key)
		return "", false
	}

	c.moveToFront(item)
	return item.value, true
}

func (c *Cache) expired(item *entry) bool {
	return !item.expiresAt.IsZero() && !time.Now().Before(item.expiresAt)
}

func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.values[key]
	if !ok {
		return false
	}

	c.remove(item)
	delete(c.values, key)
	return true
}

func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.values)
}

// Keys returns keys from most recently used to least recently used.
func (c *Cache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.values))
	for item := c.head; item != nil; item = item.next {
		keys = append(keys, item.key)
	}
	return keys
}

func (c *Cache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	for key, item := range c.values {
		if c.expired(item) {
			c.remove(item)
			delete(c.values, key)
			removed++
		}
	}
	return removed
}

func (c *Cache) StartCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.CleanupExpired()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Cache) addToFront(item *entry) {
	item.prev = nil
	item.next = c.head
	if c.head != nil {
		c.head.prev = item
	} else {
		c.tail = item
	}
	c.head = item
}

func (c *Cache) moveToFront(item *entry) {
	if item == c.head {
		return
	}
	c.remove(item)
	c.addToFront(item)
}

func (c *Cache) remove(item *entry) {
	if item.prev != nil {
		item.prev.next = item.next
	} else {
		c.head = item.next
	}

	if item.next != nil {
		item.next.prev = item.prev
	} else {
		c.tail = item.prev
	}

	item.prev = nil
	item.next = nil
}
