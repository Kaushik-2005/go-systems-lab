# Load Balancer Design

## What it solves

Clients should not need to know which backend server is available. The load balancer gives them one HTTP address and chooses a healthy backend for each request.

```text
client -> load balancer :8080 -> healthy backend
```

The request is forwarded with `httputil.ReverseProxy`, so the load balancer does not implement the backend application itself.

## Request flow

For every request:

1. collect the healthy backends;
2. return `503 Service Unavailable` if none are healthy;
3. ask the selector to choose a backend;
4. increment its active-request count;
5. apply the request timeout;
6. forward the request through the reverse proxy;
7. decrement the active-request count;
8. mark the backend unhealthy when forwarding fails.

The selector and proxy are separate. A selector decides where a request goes; the proxy handles HTTP forwarding.

## Backend state

Each `servers.Server` contains:

```text
backend URL
reverse proxy
healthy flag
active request count
```

The healthy flag uses `sync.RWMutex`. The active-request count uses `atomic.Int64` because many HTTP handlers can update it at the same time.

## Selection algorithms

All algorithms implement:

```go
Next([]*servers.Server) *Server
```

Only healthy servers are passed to the selector.

### Round robin

An index moves through the list:

```text
server-1 -> server-2 -> server-3 -> server-1
```

It is predictable and simple, but it does not consider request duration.

### Random

A random healthy backend is selected. The distribution may be uneven for a small number of requests.

### Least connections

The backend with the fewest active requests is selected. This helps when requests have different durations. Selection and increment are separate operations, so simultaneous requests can still make the choice approximate.

## Health checks

The health checker calls `GET /health` immediately and then every two seconds:

```text
start
  |
  +--> check backend 1 --+
  +--> check backend 2 ---+--> update health flags
  `--> check backend 3 --+
```

Each backend is checked in its own goroutine. A `sync.WaitGroup` waits for the checks to finish. The health-check HTTP client has a timeout, and the context stops the loop during shutdown.

An HTTP 200 response marks a backend healthy. A failed request marks it unhealthy. Healthy backends can rejoin after a later successful check.

## Failure behavior

| Situation | Result |
|---|---|
| Health check fails | The backend is excluded from selection. |
| Backend fails after selection | The proxy returns `502 Bad Gateway` and marks it unhealthy. |
| Backend does not respond before timeout | The request stops at the configured timeout and the backend is marked unhealthy. |
| All backends are unhealthy | The load balancer returns `503 Service Unavailable`. |
| Load balancer shuts down | The health-check context is cancelled and the HTTP server waits for active requests. |

## Concurrency and lifecycle

The Go HTTP server handles request concurrency. The component adds goroutines for health checks.

```text
health flag       sync.RWMutex
selector state    selector mutex
active requests   atomic.Int64
health lifecycle  context.Context
request lifetime  timeout context
```

The load balancer is one process with local backend state. It does not provide service discovery, cross-process health state, request retries, or automatic failover to another backend after a proxy error.

## Repository design

```text
algorithms/        interchangeable selection strategies
loadbalancer/      request routing and health checks
servers/           backend representation and proxy state
cmd/server/        local HTTP backend
cmd/loadbalancer/  load balancer process
```
