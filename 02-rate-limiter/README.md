# Rate Limiter

## What it is

Let's say a client is sending too many requests to your service. A rate limiter checks the client and decides whether the next request should be accepted or rejected.

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

Every algorithm has per-key state. Checking the state and updating it must happen together, otherwise concurrent requests can all observe the same old count and exceed the limit.

The diagrams below show one key at a time.

## The four algorithms

### Fixed window

Stores a window start and a counter for each key. The counter resets when the window expires. It is simple and cheap, but a client can use its full allowance at the end of one window and immediately use it again at the start of the next.

![Fixed-window timeline showing accepted requests clustered across a reset boundary](diagrams/fixed-window.svg)

Each key's window starts with its first accepted request after initialization or expiration. For a limit of 3, the first three requests in that window pass; later requests wait until the window expires.

### Sliding-window log

Stores timestamps for accepted requests. Old timestamps leave the log as time moves forward, so the remaining timestamps give an exact rolling count. The trade-off is memory usage.

![Sliding-window log showing timestamps expiring as the interval moves](diagrams/sliding-window-log.svg)

The interval moves with every request. With a 10-second window at time 25 seconds, timestamps before 15 seconds are removed. Rejected requests do not add timestamps.

### Sliding-window counter

Stores the previous and current window counts. It weights the previous count by the amount of overlap with the rolling interval. This uses constant-size state but produces an estimate.

![Sliding-window counter showing weighted overlap between adjacent windows](diagrams/sliding-window-counter.svg)

For example, 25% into the current window, the previous window has a weight of 75%. If previous is 8 and current is 2, the estimate is `8 * 0.75 + 2 = 8`. This approximates the rolling count without storing individual timestamps.

### Token bucket

Stores available tokens and the last refill time. Elapsed time adds tokens up to the capacity, and an accepted request consumes one token. Capacity controls bursts; refill rate controls sustained traffic.

![Token bucket showing a burst consuming tokens and time refilling them](diagrams/token-bucket.svg)

A bucket with capacity 5 can accept an immediate burst of five requests. At a refill rate of 2 tokens per second, an empty bucket earns one token after half a second. Refill is calculated when a request arrives; it does not need a background timer.

## Concurrency

Each limiter protects its map and check-update operation with a mutex. The limiter is local to one process, so separate processes do not share quota state.

## Project structure

```text
algorithms/       rate-limiting algorithms
ratelimiter/      shared limiter interface
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
