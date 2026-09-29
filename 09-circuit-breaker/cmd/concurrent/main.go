package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"circuit-breaker/breaker"
)

func main() {
	circuit := breaker.New(1, 100*time.Millisecond)

	var dependencyCalls atomic.Int32

	// First failure opens the circuit.
	_ = circuit.Execute(func() error {
		dependencyCalls.Add(1)
		return errors.New("dependency failed")
	})

	time.Sleep(150 * time.Millisecond)

	const callers = 10

	var group sync.WaitGroup
	start := make(chan struct{})

	results := make(chan error, callers)

	for range callers {
		group.Add(1)

		go func() {
			defer group.Done()

			<-start

			err := circuit.Execute(func() error {
				dependencyCalls.Add(1)

				time.Sleep(100 * time.Millisecond)
				return nil
			})

			results <- err
		}()
	}

	close(start)
	group.Wait()
	close(results)

	successes := 0
	rejected := 0

	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, breaker.ErrOpen) {
			rejected++
		}
	}

	fmt.Printf(
		"successes=%d rejected=%d dependency_calls=%d state=%s\n",
		successes,
		rejected,
		dependencyCalls.Load(),
		circuit.State(),
	)
}
