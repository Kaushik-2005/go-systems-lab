# Load Balancer Design

## The problem

Clients need one stable HTTP address, while the application may run on several backend servers. The load balancer chooses a healthy backend and forwards each request to it.

```text
client -> load balancer :8080 -> backend :9000
                              -> backend :9001
                              -> backend :9002
```

The load balancer does not process application logic. `httputil.ReverseProxy` forwards the request and response.

## How a request works

For every request, the component filters unhealthy backends, chooses one, increments its active-request count, applies a timeout, and forwards the request. When forwarding ends, the count is decremented. A proxy error marks that backend unhealthy and returns `502`; if no backend is healthy, the response is `503`.

The selectors implement the same behavior:

```go
Next([]*servers.Server) *Server
```

Round robin rotates through the list, random chooses a healthy backend randomly, and least connections chooses the backend with the smallest active-request count.

The health checker calls `GET /health` immediately and every two seconds. Each backend is checked in its own goroutine. A successful response marks it healthy; a failed check excludes it from routing.

## State, concurrency, and failures

Each backend stores its URL, reverse proxy, health flag, and active-request count. The health flag uses `sync.RWMutex`; the request count uses `atomic.Int64`.

The HTTP server already creates concurrent request handlers. The component adds concurrency for health checks and uses a context to stop them during shutdown.

Backend failure after selection can still happen because a health check is only a point-in-time observation. The request timeout prevents a hanging backend from blocking forever. The load balancer itself is one process, so its state is local and it is also a possible single point of failure.

## Repository design

```text
algorithms/        selection strategies
loadbalancer/      routing, proxying, and health checks
servers/           backend state and reverse proxy
cmd/server/        local backend process
cmd/loadbalancer/  load balancer process
```
