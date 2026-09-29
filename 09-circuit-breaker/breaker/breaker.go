package breaker

import (
	"errors"
	"sync"
	"time"
)

var ErrOpen = errors.New("circuit breaker is open")

type State string

const (
	Closed   State = "closed"
	Open     State = "open"
	HalfOpen State = "half-open"
)

type Breaker struct {
	mu sync.Mutex

	state        State
	failures     int
	threshold    int
	resetTimeout time.Duration
	openedAt     time.Time
	probeRunning bool
}

func New(threshold int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		state:        Closed,
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

func (b *Breaker) Execute(operation func() error) error {
	b.mu.Lock()

	if b.state == Open {
		if time.Since(b.openedAt) < b.resetTimeout {
			b.mu.Unlock()
			return ErrOpen
		}

		b.state = HalfOpen
	}

	if b.state == HalfOpen {
		if b.probeRunning {
			b.mu.Unlock()
			return ErrOpen
		}

		b.probeRunning = true
	}

	b.mu.Unlock()

	err := operation()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.failures++
		b.probeRunning = false

		if b.state == HalfOpen || b.failures >= b.threshold {
			b.state = Open
			b.openedAt = time.Now()
		}

		return err
	}

	b.failures = 0
	b.probeRunning = false
	b.state = Closed

	return nil
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.state
}
