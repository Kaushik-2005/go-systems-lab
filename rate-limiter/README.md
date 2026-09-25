# Rate Limiter

## What is a rate limiter?

A rate limiter controls how often a client, user, or resource may perform an operation during a period of time.

```text
request -> rate limiter -> accepted or rejected
                       |
                       +-- key: client/user/resource
                       +-- algorithm: fixed window, sliding window, token bucket
```

This component implements request-admission algorithms from scratch in Go. It progresses from a fixed-window counter to more precise and flexible algorithms.

## Learning sequence

1. Fixed-window counter.
2. Sliding-window log.
3. Sliding-window counter.
4. Token bucket.

## How the algorithms work

### Fixed window

One counter is used for each key inside a fixed time interval.

```text
time  ─────────────────────────────────────────────>

window 1:       |---------------- 5 seconds ----------------|
client-1:       request request request   reject   reject
                count=1 count=2 count=3   limit reached

window 2:       |---------------- 5 seconds ----------------|
client-1:       counter resets to zero
```

The reset makes the algorithm simple, but it can allow a burst at the boundary between two windows.

### Sliding-window log

Each key stores timestamps for its accepted requests. Older timestamps leave the log as time moves forward.

```text
current time:                         now
window:                 |------------- 5 seconds -------------|
timestamps:                    t1       t2       t3       now
                               keep     keep     keep     check

old timestamp < now - 5s  ---> removed
```

The count is accurate for the rolling interval, but storing timestamps uses more memory.

### Sliding-window counter

The counter approximates the rolling interval using two neighboring fixed-window counts.

```text
previous window              current window
|--------------------|       |--------------------|
        count = 3                     count = 1
                    ^
                    | current time

estimated count = previous count × overlap + current count
```

It uses constant-size state per key, but it estimates rather than stores exact request times.

### Token bucket

Tokens refill over time. Each accepted request consumes one token.

```text
                 refill: 1 token/second
                          ↓
                    +-----------+
                    |  tokens   |  capacity = 3
                    |  ● ● ●    |
                    +-----+-----+
                          |
                          | request consumes 1 token
                          v
                     accepted request
```

The bucket allows an initial burst up to its capacity, then limits sustained traffic according to the refill rate.

## Project structure

```text
algorithms/       limiting strategies
cmd/              runnable demonstration
ratelimiter/      shared limiter interfaces and state
design.md         design reasoning and trade-offs
go.mod            Go module definition
```

## Current status

All four planned algorithms are implemented. Each algorithm supports independent per-key state and protects its check-and-update operation for concurrent callers.

Run the demonstration from the `rate-limiter` directory:

```powershell
go run ./cmd/fixedwindow -limit 3 -window 5s -requests 5
```

The first three requests should be allowed and the remaining requests should be rejected until the window expires.

Run the sliding-window log with:

```powershell
go run ./cmd/fixedwindow -algorithm sliding -limit 3 -window 5s -requests 5
```

To compare it with the fixed-window boundary experiment, send the second batch before the five-second window has fully elapsed:

```powershell
go run ./cmd/fixedwindow -algorithm sliding -limit 3 -window 5s -requests 3 -batches 2 -wait 4.9s
```

The second batch should be rejected because the first batch's timestamps are still inside the rolling five-second window.

Run the sliding-window counter with:

```powershell
go run ./cmd/fixedwindow -algorithm sliding-counter -limit 3 -window 5s -requests 5
```

The counter uses less memory than the sliding log, but its result is an estimate based on adjacent-window counts.

Run the token bucket with a capacity of three tokens and a refill rate of one token per second:

```powershell
go run ./cmd/fixedwindow -algorithm token -capacity 3 -refill-rate 1 -requests 5
```

The first three requests can use the initial burst. Later requests are rejected until tokens refill.

To observe the fixed-window boundary burst, send two full batches just over one window apart:

```powershell
go run ./cmd/fixedwindow -limit 3 -window 5s -requests 3 -batches 2 -wait 5.1s
```

Both batches are allowed, so six requests can be accepted within only a little more than five seconds. This boundary behavior motivates the sliding-window log.

## Example output explained

For the default fixed-window command:

```text
time=14:00:00.000 batch=1 request=1 key=client-1 allowed=true
time=14:00:00.001 batch=1 request=2 key=client-1 allowed=true
time=14:00:00.001 batch=1 request=3 key=client-1 allowed=true
time=14:00:00.001 batch=1 request=4 key=client-1 allowed=false
```

- `time` is when the demo called `Limiter.Allow`.
- `batch` identifies a group of requests separated by the configured `-wait` duration.
- `request` is the request number in the demo, not a network request ID.
- `key=client-1` is the per-key state being limited.
- `allowed=true` means the algorithm accepted the request and updated its state.
- `allowed=false` means the limit was reached for the relevant window or bucket state.

The same command shape runs all four algorithms. Only the internal state calculation changes: fixed window uses one counter, sliding log uses timestamps, sliding counter uses adjacent-window counts, and token bucket uses refillable tokens.
