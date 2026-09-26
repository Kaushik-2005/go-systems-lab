# Rate Limiter Design

## What it solves

A rate limiter protects work by deciding whether a request may proceed:

```text
Limiter.Allow(key)
       |
       +--> true  -> perform the operation
       `--> false -> reject the operation
```

The key identifies who owns the quota. Different clients have different state, so one client's requests do not consume another client's allowance.

All algorithms implement:

```go
Allow(key string) bool
```

The caller uses the same interface while the internal counting method changes.

## The important concurrency rule

Checking the limit and updating the state must be one protected operation.

Without that rule, two goroutines could both see a count of two, both decide that a limit of three allows them, and both update the count. The result would be more accepted requests than the limit permits.

Each limiter protects its per-key map and check-update operation with a mutex.

## Algorithm designs

### Fixed window

State per key:

```text
window start + accepted count
```

On each request:

1. load or create the key state;
2. reset it if the fixed window expired;
3. reject when the count reaches the limit;
4. otherwise increment and accept.

The algorithm is inexpensive, but a client can use its full allowance at the end of one window and again at the beginning of the next.

### Sliding-window log

State per key:

```text
timestamps of accepted requests
```

Old timestamps are removed on every request. The remaining timestamps are the exact accepted requests inside the rolling interval. This is accurate but uses memory proportional to recent traffic.

### Sliding-window counter

State per key:

```text
previous count + current count + current window start
```

The previous count is weighted by the part of that window that overlaps the rolling interval:

```text
estimated requests = previous count * overlap + current count
```

It uses constant-size state and gives an estimate rather than an exact timestamp count.

### Token bucket

State per key:

```text
tokens available + last refill time
```

On each request:

1. calculate tokens earned since the last request;
2. cap the bucket at its capacity;
3. reject if fewer than one token is available;
4. consume one token and accept otherwise.

Capacity controls burst size. Refill rate controls the long-term rate.

## Repository design

```text
ratelimiter/      Limiter interface and shared definitions
algorithms/       four state and decision implementations
cmd/fixedwindow/  command-line experiments
```

The command supplies algorithm parameters and prints decisions. The algorithm packages own the maps, timestamps, counters, and token state.

## Demonstrations

The fixed-window boundary experiment sends two batches just over one window apart. Both batches can pass, which demonstrates the boundary burst.

The sliding-log experiment sends the second batch before the rolling window expires. The earlier timestamps are still present, so the second batch is rejected.

The token-bucket experiment begins with a burst equal to the bucket capacity. Later requests are accepted only as tokens refill.

## Guarantees and scope

The implementation provides per-key, synchronized admission decisions inside one process. State disappears when the process stops, and separate processes do not share quotas. The sliding-window counter is intentionally approximate. The package is a limiter mechanism, not HTTP middleware or a distributed rate-limiting service.
