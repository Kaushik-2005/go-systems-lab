# Load Balancer

## What is a load balancer?

A load balancer is a system that receives requests at one stable address and distributes them across multiple backend servers.

It helps prevent one backend from handling all the traffic, allows work to be shared, and can stop sending requests to a backend that is unavailable.

```text
                             +--> server-1
                             |
client --> load balancer    -+--> server-2
                             |
                             +--> server-3
```

The load balancer decides which backend should handle each request. That decision is called a selection strategy. Real systems may also perform health checks, enforce timeouts, retry failed requests, and shut down gracefully.

## What we implemented

This component is a small HTTP load balancer built from Go's standard library. It accepts client requests, selects a healthy backend, and forwards the request through an HTTP reverse proxy.

## Concepts

- HTTP reverse proxying with `net/http/httputil`
- interchangeable backend-selection algorithms
- round robin, random, and least connections
- active health checks and recovery
- concurrent shared state with mutexes and atomics
- request timeouts and backend failure handling
- graceful shutdown with `context.Context`

## Architecture

```text
                        +----------------------+
                        |   health checker     |
                        |   GET /health        |
                        +----------+-----------+
                                   |
                                   v
                 +-----------------+-----------------+
+--------+       |      load balancer :8080          |
| client | ----> |                                   |
+--------+       |    selector + request timeout      |
                 +-----------------+-----------------+
                                   |
                 +-----------------+-----------------+
                 |                 |                 |
                 v                 v                 v
           +-----------+     +-----------+     +-----------+
           | backend 1 |     | backend 2 |     | backend 3 |
           |  :9000    |     |  :9001    |     |  :9002    |
           +-----------+     +-----------+     +-----------+
```

## Project structure

```text
algorithms/        selection strategies
cmd/loadbalancer/  load balancer executable
cmd/server/        local backend executable
loadbalancer/      request routing and health checking
servers/           backend state and reverse proxy
design.md          design reasoning and limitations
go.mod             Go module definition
```

## Running locally

Run these commands from the `load-balancer` directory. Open separate terminals and start three backends:

```powershell
$env:PORT="9000"; $env:SERVER_ID="server-1"; go run ./cmd/server
```

```powershell
$env:PORT="9001"; $env:SERVER_ID="server-2"; go run ./cmd/server
```

```powershell
$env:PORT="9002"; $env:SERVER_ID="server-3"; go run ./cmd/server
```

Start the load balancer in another terminal:

```powershell
go run ./cmd/loadbalancer -algorithm roundrobin -backends http://localhost:9000,http://localhost:9001,http://localhost:9002
```

Send a request:

```powershell
(Invoke-WebRequest http://localhost:8080/test -UseBasicParsing).Content
```

The backend selection algorithm can be changed with `-algorithm`:

```text
roundrobin          predictable rotation
random              random healthy backend
leastconnections    backend with the fewest active requests
```

The default request timeout is five seconds. For failure experiments, use `-request-timeout 2s`.

The backend exposes `GET /health`. The load balancer checks it immediately and every two seconds. Press `Ctrl+C` in the load balancer terminal for graceful shutdown.

## Demonstrations

To make least-connections behavior visible, start a backend with an artificial delay:

```powershell
$env:PORT="9000"; $env:SERVER_ID="server-1"; $env:DELAY_MS="1000"; go run ./cmd/server
```

Then send concurrent requests with PowerShell 7:

```powershell
1..20 | ForEach-Object -Parallel {
    (Invoke-WebRequest http://localhost:8080/test -UseBasicParsing).Content
} -ThrottleLimit 20
```

### Example output explained

```text
server-1 handled GET /test
server-2 handled GET /test
server-3 handled GET /test
server-1 handled GET /test
```

- `server-1`, `server-2`, and `server-3` are the backend identities from `SERVER_ID`.
- `handled` confirms the request reached a backend rather than ending at the load balancer.
- `GET` is the HTTP method forwarded by `httputil.ReverseProxy`.
- `/test` is the original request path preserved during forwarding.
- The repeating order comes from `algorithms.RoundRobin.Next` selecting the next healthy backend.

If one backend is stopped, later output contains only the remaining server IDs after the health checker marks the stopped backend unhealthy. With `-algorithm random`, the order varies. With `-algorithm leastconnections`, concurrent requests are routed using each backend's active-request count.

## Implemented

- single-backend reverse proxy
- multiple backend representation
- round-robin selection
- random selection
- least-connections selection
- active and concurrent health checks
- unhealthy backend skipping and recovery
- request timeout and immediate backend failure handling
- graceful process shutdown
