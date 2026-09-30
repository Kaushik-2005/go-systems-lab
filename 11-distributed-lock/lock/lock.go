package lock

import (
	"sync"
	"time"
)

type Lock struct {
	mu        sync.Mutex
	held      bool
	owner     string
	token     uint64
	nextToken uint64
	expiresAt time.Time
}

func New() *Lock {
	return &Lock{}
}

func (l *Lock) Acquire(owner string, ttl time.Duration) (uint64, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.expireLocked()

	if l.held {
		return 0, false
	}

	l.nextToken++
	l.held = true
	l.owner = owner
	l.token = l.nextToken
	l.expiresAt = time.Now().Add(ttl)

	return l.token, true
}

func (l *Lock) Renew(owner string, token uint64, ttl time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.expireLocked()

	if !l.held || l.owner != owner || l.token != token {
		return false
	}

	l.expiresAt = time.Now().Add(ttl)
	return true
}

func (l *Lock) Release(owner string, token uint64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.expireLocked()

	if !l.held || l.owner != owner || l.token != token {
		return false
	}

	l.held = false
	l.owner = ""
	l.token = 0
	l.expiresAt = time.Time{}

	return true
}

func (l *Lock) expireLocked() {
	if l.held && time.Now().After(l.expiresAt) {
		l.held = false
		l.owner = ""
		l.token = 0
		l.expiresAt = time.Time{}
	}
}
