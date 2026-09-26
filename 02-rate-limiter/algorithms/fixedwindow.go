package algorithms

import (
	"sync"
	"time"
)

type windowState struct {
	start time.Time
	count int
}

// FixedWindow allows a maximum number of operations per time window and key.
type FixedWindow struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	states map[string]windowState
}

func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:  limit,
		window: window,
		states: make(map[string]windowState),
	}
}

func (f *FixedWindow) Allow(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	state, exists := f.states[key]
	if !exists || now.Sub(state.start) >= f.window {
		f.states[key] = windowState{start: now, count: 1}
		return true
	}

	if state.count >= f.limit {
		return false
	}

	state.count++
	f.states[key] = state
	return true
}
