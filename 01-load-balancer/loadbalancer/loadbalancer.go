package loadbalancer

import (
	"context"
	"net/http"
	"time"

	"load-balancer/algorithms"
	"load-balancer/servers"
)

type LoadBalancer struct {
	backends []*servers.Server
	selector algorithms.Selector
	timeout  time.Duration
}

func New(backends []*servers.Server, selector algorithms.Selector, timeout time.Duration) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
		selector: selector,
		timeout:  timeout,
	}
}

func (lb *LoadBalancer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	healthyBackends := make([]*servers.Server, 0, len(lb.backends))
	for _, backend := range lb.backends {
		if backend.IsHealthy() {
			healthyBackends = append(healthyBackends, backend)
		}
	}

	backend := lb.selector.Next(healthyBackends)
	if backend == nil {
		http.Error(writer, "no backends configured", http.StatusServiceUnavailable)
		return
	}

	backend.StartRequest()
	defer backend.FinishRequest()

	requestContext, cancel := context.WithTimeout(request.Context(), lb.timeout)
	defer cancel()
	request = request.WithContext(requestContext)

	backend.Proxy.ServeHTTP(writer, request)
}
