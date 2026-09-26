package algorithms

import (
	"sync"
	"time"
)

type counterState struct {
	windowStart time.Time
	previous    int
	current     int
}

// SlidingWindowCounter estimates recent traffic using two adjacent windows.
type SlidingWindowCounter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	states map[string]counterState
}

func NewSlidingWindowCounter(limit int, window time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		limit:  limit,
		window: window,
		states: make(map[string]counterState),
	}
}

func (s *SlidingWindowCounter) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	state, exists := s.states[key]
	if !exists {
		state = counterState{windowStart: now.Truncate(s.window)}
	}

	elapsed := now.Sub(state.windowStart)
	if elapsed >= s.window {
		windowsPassed := int(elapsed / s.window)
		if windowsPassed == 1 {
			state.previous = state.current
		} else {
			state.previous = 0
		}
		state.current = 0
		state.windowStart = state.windowStart.Add(time.Duration(windowsPassed) * s.window)
		elapsed = now.Sub(state.windowStart)
	}

	previousWeight := 1 - float64(elapsed)/float64(s.window)
	estimated := float64(state.previous)*previousWeight + float64(state.current)
	if estimated >= float64(s.limit) {
		s.states[key] = state
		return false
	}

	state.current++
	s.states[key] = state
	return true
}
