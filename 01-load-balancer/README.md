# Load Balancer

## What it is

Lets say your application is running on multiple servers. You don't want the client to choose a server or send every request to one server, so a load balancer receives the request and chooses a backend for it.

```text
client -> load balancer -> backend server
                       -> backend server
                       -> backend server
```

## How it works

The load balancer checks which backends are healthy, selects one, and forwards the request through an HTTP reverse proxy.

It also handles:

- round-robin, random, and least-connections selection;
- active health checks through `GET /health`;
- request timeouts;
- failed backend handling;
- graceful shutdown.

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
