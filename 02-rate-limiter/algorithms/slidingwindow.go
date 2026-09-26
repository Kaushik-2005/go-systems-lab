package algorithms

import (
	"sync"
	"time"
)

// SlidingWindowLog stores accepted request timestamps for each key.
type SlidingWindowLog struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	logs   map[string][]time.Time
}

func NewSlidingWindowLog(limit int, window time.Duration) *SlidingWindowLog {
	return &SlidingWindowLog{
		limit:  limit,
		window: window,
		logs:   make(map[string][]time.Time),
	}
}

func (s *SlidingWindowLog) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-s.window)
	requests := s.logs[key]

	firstRecent := 0
	for firstRecent < len(requests) && requests[firstRecent].Before(cutoff) {
		firstRecent++
	}
	requests = requests[firstRecent:]

	if len(requests) >= s.limit {
		s.logs[key] = requests
		return false
	}

	s.logs[key] = append(requests, now)
	return true
}
