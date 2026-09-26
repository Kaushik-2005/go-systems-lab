# Load Balancer

## What is a load balancer?

A load balancer gives clients one address and spreads their requests across multiple backend servers.

```text
client
   |
   v
load balancer :8080
   |       |       |
   v       v       v
server-1 server-2 server-3
 :9000    :9001    :9002
```

This prevents one server from receiving all the traffic. It also gives the system a place to check backend health and stop routing to a server that is unavailable.

## How this implementation works

For each request, the load balancer:

1. keeps only healthy backends;
2. chooses one with the configured algorithm;
3. forwards the request through an HTTP reverse proxy;
4. tracks the active request;
5. marks the backend unhealthy if forwarding fails.

The health checker calls `GET /health` on every backend. Requests have a timeout, and the process shuts down through `context.Context`.

## Selection algorithms

```text
roundrobin       server-1 -> server-2 -> server-3 -> server-1
random           chooses a random healthy backend
leastconnections chooses the backend with the fewest active requests
```

## Project structure

```text
algorithms/        round robin, random, and least connections
cmd/loadbalancer/  load balancer executable
cmd/server/        local backend executable
loadbalancer/      routing, proxying, and health checks
servers/           backend state and reverse proxy
go.mod
design.md          routing, concurrency, and failure behavior
```

## Run it

Run these commands from the `01-load-balancer` directory in separate terminals:

```powershell
$env:PORT="9000"; $env:SERVER_ID="server-1"; go run ./cmd/server
```

```powershell
$env:PORT="9001"; $env:SERVER_ID="server-2"; go run ./cmd/server
```

```powershell
$env:PORT="9002"; $env:SERVER_ID="server-3"; go run ./cmd/server
```

Start the load balancer:

```powershell
go run ./cmd/loadbalancer -algorithm roundrobin -backends http://localhost:9000,http://localhost:9001,http://localhost:9002
```

Send requests:

```powershell
(Invoke-WebRequest http://localhost:8080/test -UseBasicParsing).Content
```

Choose another algorithm with `-algorithm random` or `-algorithm leastconnections`.

## Health and failure demonstration

Stop one backend and wait for the next health-check interval. Requests then use the remaining healthy servers. Restart the backend and it can rejoin after a successful health check.

To make least-connections behavior visible, add a delay to a backend:

```powershell
$env:PORT="9000"; $env:SERVER_ID="server-1"; $env:DELAY_MS="1000"; go run ./cmd/server
```

Then send concurrent requests with PowerShell 7:

```powershell
1..20 | ForEach-Object -Parallel {
    (Invoke-WebRequest http://localhost:8080/test -UseBasicParsing).Content
} -ThrottleLimit 20
```

The response identifies the backend that handled each request. For example, `server-2 handled GET /test` means the request passed through the load balancer and was served by server 2.
