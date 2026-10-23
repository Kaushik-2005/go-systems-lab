# Load Balancer

## What it is

Let's say your application is running on multiple servers. You don't want the client to choose a server or send every request to one server, so a load balancer receives the request and chooses a backend for it.

![Request routing through healthy backends, with an unhealthy backend excluded](diagrams/request-routing.svg)

## How it works

The load balancer checks which backends are healthy, selects one, and forwards the request through an HTTP reverse proxy.

It also handles:

- round-robin, random, and least-connections selection;
- active health checks through `GET /health`;
- request timeouts;
- failed backend handling;
- graceful shutdown.

### Request lifecycle

For every request, the component filters unhealthy backends, chooses one, increments its active-request count, applies a timeout, and forwards the request with `httputil.ReverseProxy`. When forwarding ends, the count is decremented. A proxy error marks that backend unhealthy and returns `502`; if no backend is healthy, the response is `503`.

### Selection algorithms

The selectors implement a shared interface:

```go
type Selector interface {
    Next([]*servers.Server) *servers.Server
}
```

![Round-robin, random, and least-connections selection compared on three healthy backends](diagrams/selection-strategies.svg)

All three algorithms receive only healthy backends. If the list is empty, they return `nil` and the load balancer responds with `503`.

#### Round robin

Round robin gives each healthy backend a turn. It keeps an index, selects `backends[index % len(backends)]`, then increments the index. Modulo wraps selection back to the beginning of the list.

![Six round-robin requests cycle through A, B, C, A, B, C as the index increases](diagrams/round-robin.svg)

For a stable list `[A, B, C]`, requests go to `A, B, C, A, B, C`. A mutex protects the index so concurrent selections do not lose updates. When health changes, the same index is applied to the new healthy list; the sequence can change.

Selection takes O(1) time. This works well when backends have similar capacity and requests have similar cost. It distributes request counts evenly for a stable list, but does not account for a backend already processing slow requests.

Use `-algorithm roundrobin`.

#### Random

Random selection draws an integer with `rng.Intn(len(backends))` and uses it as the backend index. With three healthy backends, each has a one-third chance of being selected on every request.

![Independent random draws can route consecutive requests to the same backend](diagrams/random-selection.svg)

An example sequence is `B, B, A, C, A, C`; actual results vary. Selecting B once does not prevent it from being selected again immediately. A mutex protects the shared random-number generator during concurrent selections.

Selection takes O(1) time. Traffic tends to spread evenly over many requests, but a short burst can concentrate on one backend. Like round robin, it does not inspect how busy a backend is.

Use `-algorithm random`.

#### Least connections

Least connections scans all healthy backends and selects the one with the smallest active-request count. In this implementation, that count measures requests currently being proxied, rather than open TCP connections.

![Backend B is selected because its one active request is fewer than A's four and C's three](diagrams/least-connections.svg)

If A has 4 active requests, B has 1, and C has 3, the next request selects B. Before proxying, `StartRequest` increments B's count to 2. A deferred `FinishRequest` decrements it when forwarding finishes, including on failure. The next selection reads the counters again.

Ties favor the first backend in the healthy list. The selector uses a mutex and each counter uses atomic operations, but selection and incrementing the chosen backend's count are separate operations. Concurrent requests can therefore select the same backend before either increments its count.

Selection takes O(n) time for n healthy backends. It adapts to requests with different durations, since slow requests keep their backend's count elevated. The count is still only an estimate of load: it does not measure CPU usage, memory pressure, or backend capacity.

Use `-algorithm leastconnections`.

### Health checks and recovery

![Health probes mark backends healthy on HTTP 200 and unhealthy on errors or other responses](diagrams/health-checks.svg)

The command starts health checks immediately and every two seconds. Each backend is checked in its own goroutine with a two-second HTTP client timeout. A successful `GET /health` response must have status `200`; a failed check excludes the backend from routing. Later successful checks restore it.

## State, concurrency, and failures

Each backend stores its URL, reverse proxy, health flag, and active-request count. The health flag uses `sync.RWMutex`; the request count uses `atomic.Int64`.

The HTTP server runs request handlers concurrently. Health checks use their own goroutines and a context to stop scheduling checks during shutdown. The command handles interrupt and termination signals, then allows up to five seconds for the HTTP server to shut down gracefully.

Backend failure after selection can still happen because a health check is only a point-in-time observation. The request timeout prevents a hanging backend from blocking forever. The load balancer itself is one process, so its state is local and it is also a possible single point of failure.

## Project structure

```text
algorithms/        backend selection algorithms
loadbalancer/      routing, proxying, and health checks
servers/           backend state and reverse proxy
cmd/server/        local backend server
cmd/loadbalancer/  load balancer process
go.mod
```

## Run it

Run these commands from the `01-load-balancer` directory in separate terminals:

```powershell
$env:PORT="9000"; $env:SERVER_ID="server-1"; go run ./cmd/server
$env:PORT="9001"; $env:SERVER_ID="server-2"; go run ./cmd/server
$env:PORT="9002"; $env:SERVER_ID="server-3"; go run ./cmd/server
```

Start the load balancer:

```powershell
go run ./cmd/loadbalancer -algorithm roundrobin -backends http://localhost:9000,http://localhost:9001,http://localhost:9002
```

Send a request:

```powershell
(Invoke-WebRequest http://localhost:8080/test -UseBasicParsing).Content
```

Stop one backend and send more requests. After the health checker detects the failure, traffic moves to the remaining healthy backends.
