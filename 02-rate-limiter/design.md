# Rate Limiter Design

## The problem

A service needs a simple decision before doing work: should this request be allowed? The key identifies the quota owner, such as a client, user, API token, or resource.

```go
Allow(key string) bool
```

Every algorithm has per-key state. Checking the state and updating it must happen together, otherwise concurrent requests can all observe the same old count and exceed the limit.

## The four algorithms

### Fixed window

Stores a window start and a counter for each key. The counter resets when the window expires. It is simple and cheap, but a client can use its full allowance at the end of one window and immediately use it again at the start of the next.

### Sliding-window log

Stores timestamps for accepted requests. Old timestamps leave the log as time moves forward, so the remaining timestamps give an exact rolling count. The trade-off is memory usage.

### Sliding-window counter

Stores the previous and current window counts. It weights the previous count by the amount of overlap with the rolling interval. This uses constant-size state but produces an estimate.

### Token bucket

Stores available tokens and the last refill time. Elapsed time adds tokens up to the capacity, and an accepted request consumes one token. Capacity controls bursts; refill rate controls sustained traffic.

## Concurrency and repository design

Each limiter protects its map and check-update operation with a mutex. The limiter is local to one process, so separate processes do not share quota state.

```text
ratelimiter/      shared Limiter interface
algorithms/       four limiting implementations
cmd/fixedwindow/  configuration and runnable demonstrations
```

The demos make the differences visible: fixed-window boundary bursts, exact sliding-log rejection, sliding-counter estimation, and token-bucket refill behavior.
