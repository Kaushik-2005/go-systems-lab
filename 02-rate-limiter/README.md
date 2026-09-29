# Rate Limiter

## What it is

Lets say a client is sending too many requests to your service. A rate limiter checks the client and decides whether the next request should be accepted or rejected.

```text
request -> rate limiter -> accepted
                       -> rejected
```

## How it works

Each key gets its own limiting state. A key can represent a client, user, API token, or resource.

The component implements four algorithms:

- fixed-window counter;
- sliding-window log;
- sliding-window counter;
- token bucket.

The algorithms share the same `Allow(key)` decision but calculate the limit differently.

## Project structure

```text
algorithms/       rate-limiting algorithms
ratelimiter/      shared limiter interface and state
cmd/fixedwindow/  runnable demonstrations
go.mod
```

## Run it

Run from the `02-rate-limiter` directory:

```powershell
go run ./cmd/fixedwindow -algorithm fixed -limit 3 -window 5s -requests 5
go run ./cmd/fixedwindow -algorithm sliding -limit 3 -window 5s -requests 5
go run ./cmd/fixedwindow -algorithm sliding-counter -limit 3 -window 5s -requests 5
go run ./cmd/fixedwindow -algorithm token -capacity 3 -refill-rate 1 -requests 5
```

The fixed window uses a counter, the sliding log stores timestamps, the sliding counter estimates the rolling count, and the token bucket refills tokens over time.
