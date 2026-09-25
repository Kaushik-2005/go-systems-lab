# Load Balancer Design

## 1. System overview

The load balancer gives clients one stable HTTP address while distributing requests across multiple backend servers.

```text
client
  |
  v
load balancer :8080
  |-- filter healthy backends
  |-- select one backend
  |-- apply request timeout
  `-- reverse proxy the request
  |
  +--> backend :9000
  +--> backend :9001
  `--> backend :9002
```

The load balancer itself does not process the application request. It chooses a backend and forwards the request and response through `httputil.ReverseProxy`.

## 2. Request decision model

For every request, the load balancer performs these steps:

1. build a list of currently healthy backends;
2. return `503 Service Unavailable` if the list is empty;
3. ask the configured selector to choose one backend;
4. increment that backend's active-request count;
5. attach the configured request timeout;
6. forward the request through the backend's reverse proxy;
7. decrement the active-request count when forwarding returns;
8. mark the backend unhealthy if forwarding fails.

The selector does not know how requests are proxied. The proxy does not know how a backend was selected. This keeps the algorithm and networking mechanisms separate.

## 3. Backend state

`servers.Server` owns the state associated with one backend:

```text
backend URL
reverse proxy
healthy flag
active request count
```

The healthy flag is protected by `sync.RWMutex`. The active-request count uses `atomic.Int64` because it is incremented and decremented frequently by concurrent handlers.

Each backend also exposes `GET /health` through the demo server. A successful HTTP 200 response means the backend is currently considered healthy.

## 4. Selection algorithms

All selectors implement:

```go
Next([]*servers.Server) *servers.Server
```

Only healthy backends are passed to this method.

### Round robin

Stores an index and visits backends in order:

```text
backend-1 → backend-2 → backend-3 → backend-1 → ...
```

It is predictable but does not account for request duration or backend capacity.

### Random

Chooses an index randomly from the healthy backend list. It can distribute traffic over time, but equal distribution is not guaranteed for a small sample.

### Least connections

Chooses the backend with the smallest active-request count. It is useful when requests have different durations, but selection and increment are separate operations in this educational implementation, so extreme concurrent arrivals can briefly make an imperfect choice.

## 5. Health-check lifecycle

The load balancer probes every backend's `/health` endpoint immediately and then every two seconds.

```text
start
  |
  +--> probe backend-1 --+
  +--> probe backend-2 ---+--> update health states
  `--> probe backend-3 --+
          concurrently
```

Each cycle starts one goroutine per backend and waits with a `sync.WaitGroup`. A two-second HTTP client timeout prevents a health probe from waiting forever.

- failed probe: mark backend unhealthy;
- successful probe: mark backend healthy;
- request routing: exclude unhealthy backends;
- process cancellation: stop the health-check loop.

## 6. Failure handling

### Backend is unavailable during a health check

The failed probe marks it unhealthy. Future requests exclude it. If it returns and `/health` succeeds, it can rejoin.

### Backend fails after selection

The reverse proxy's `ErrorHandler` marks the selected backend unhealthy and returns `502 Bad Gateway`.

### Backend hangs

Each forwarded request receives a context timeout, five seconds by default. When the timeout expires, the proxy reports an error, marks the backend unhealthy, and returns an error instead of waiting indefinitely.

### All backends are unhealthy

The load balancer returns `503 Service Unavailable`. It does not queue requests or invent a replacement backend.

### Load balancer shuts down

`signal.NotifyContext` cancels the health-check loop. `http.Server.Shutdown` stops new connections and waits up to five seconds for active requests.

## 7. Concurrency model

The Go HTTP server owns request-handler concurrency. The component adds concurrency only for health checks.

Shared state is protected as follows:

```text
health flag          sync.RWMutex
selector state       mutex inside each selector
active requests      atomic.Int64
health lifecycle     context.Context
request lifetime     derived timeout context
```

The health checker writes backend health while request handlers read it. The locking rule keeps those operations race-safe.

## 8. Manual demonstrations

Run three backend servers and the load balancer from the `load-balancer` directory. With round robin, repeated requests rotate through the backends.

With `-algorithm random`, repeated requests arrive in varying order.

For least-connections, start backends with `DELAY_MS=1000` and send concurrent requests. The delay keeps requests active long enough for the active-request counts to influence selection.

Stop one backend and wait for the next health-check interval. Requests then use only the remaining healthy backends. Restarting the backend allows it to rejoin after a successful probe.

For immediate failure handling, delay a backend for ten seconds and run the load balancer with `-request-timeout 2s`. A request routed to that backend returns `502 Bad Gateway` and the backend is marked unhealthy.

Press `Ctrl+C` in the load balancer terminal to verify clean shutdown.

## 9. Current behavior and limitations

- health is based on a single HTTP status check;
- there is no health-check backoff, hysteresis, or failure threshold;
- a backend may fail immediately after a successful health check;
- least-connections selection is approximate under extreme concurrency;
- the load balancer itself is a single point of failure;
- backend state is local to one load-balancer process;
- there is no persistent state, service discovery, or cross-process coordination.
