package loadbalancer

import (
	"context"
	"net/http"
	"sync"
	"time"
)

func (lb *LoadBalancer) StartHealthChecks(ctx context.Context, interval time.Duration) {
	client := &http.Client{Timeout: 2 * time.Second}

	check := func() {
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(lb.backends))

		for _, backend := range lb.backends {
			go func() {
				defer waitGroup.Done()
				backend.SetHealthy(backend.CheckHealth(client))
			}()
		}

		waitGroup.Wait()
	}

	go func() {
		check()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				check()
			case <-ctx.Done():
				return
			}
		}
	}()
}
