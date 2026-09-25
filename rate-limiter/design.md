# Rate Limiter Design

## 1. System overview

A rate limiter sits before protected work and decides whether a request may proceed.

```text
request
   |
   v
Limiter.Allow(key)
   |
   +--> true  -> perform the operation
   |
   `--> false -> reject the request
```

The key identifies the independent quota owner. It may represent a client, user, API token, or resource.

All algorithms implement the same behavior:

```go
Allow(key string) bool
```

This keeps the admission decision separate from the algorithm used to calculate it.

## 2. Request decision model

For every request, an algorithm must answer three questions:

1. Which state belongs to this key?
2. Is that state still within the relevant time range?
3. If the request is accepted, how is the state updated?

The check and update must happen as one protected operation. Otherwise concurrent requests can observe the same old count and exceed the limit together.

## 3. Algorithm progression

### Fixed-window counter

State per key:

```text
window start + accepted request count
```

When a request arrives:

1. create or load the key's state;
2. reset the state if the fixed window expired;
3. reject when the count reaches the limit;
4. otherwise increment the count and accept.

This is simple and cheap, but a client can use its full limit at the end of one window and immediately use it again at the start of the next.

### Sliding-window log

State per key:

```text
timestamps of accepted requests
```

When a request arrives, timestamps older than `now - window` are removed. The remaining timestamps represent the exact recent request history. The request is accepted only when their count is below the limit.

This removes the fixed-window boundary burst, but memory usage grows with the number of recent accepted requests.

### Sliding-window counter

State per key:

```text
previous fixed-window count
current fixed-window count
current window start
```

The previous count is weighted by the fraction of that window that overlaps the current rolling interval:

```text
estimated recent requests =
    previous count × overlap fraction + current count
```

This uses constant-size state, but it is an estimate because individual request timestamps are not stored.

### Token bucket

State per key:

```text
current token count + last refill time
```

When a request arrives:

1. add tokens according to elapsed time;
2. cap the bucket at its capacity;
3. reject if fewer than one token is available;
4. otherwise consume one token and accept.

Capacity controls the allowed burst size. Refill rate controls the sustained rate.

## 4. State ownership

The `ratelimiter` package defines the shared `Limiter` interface. The `algorithms` package owns each algorithm's state and decision logic. The demonstration command only supplies configuration and prints decisions; it does not contain limiting rules.

Each key has independent state. One client's traffic does not consume another client's quota.

## 5. Concurrency model

The current implementations protect their maps and check-update operations with one mutex per limiter instance. This makes each `Allow` call atomic from the caller's perspective.

The mutex protects:

- lookup or creation of per-key state;
- expiration and cleanup;
- limit comparison;
- state mutation after an accepted request.

The implementation favors clear ownership over lock-free optimization. The limiter is local to one process; no state is shared between multiple processes.

## 6. Manual demonstrations

Run the fixed-window boundary experiment:

```powershell
go run ./cmd/fixedwindow -algorithm fixed -limit 3 -window 5s -requests 3 -batches 2 -wait 5.1s
```

Both batches are accepted, showing the boundary burst.

Run the equivalent sliding-log experiment:

```powershell
go run ./cmd/fixedwindow -algorithm sliding -limit 3 -window 5s -requests 3 -batches 2 -wait 4.9s
```

The second batch is rejected because the first batch is still inside the rolling interval.

Run the token-bucket burst experiment:

```powershell
go run ./cmd/fixedwindow -algorithm token -capacity 3 -refill-rate 1 -requests 5
```

The initial three requests consume the burst capacity; later requests wait for refill.

## 7. Current behavior and limitations

- state exists only in memory and disappears when the process stops;
- there is no distributed or shared rate limit across processes;
- expired per-key state is cleaned lazily when that key is used again;
- the sliding-window counter is approximate by design;
- the demonstration is a command-line program, not an HTTP middleware layer yet.
