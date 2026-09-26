# Rate Limiter

## What is a rate limiter?

A rate limiter decides whether an operation should be accepted or rejected based on how often a key has made requests.

```text
request + key
     |
     v
rate limiter
   |       |
 allow   reject
```

The key can represent a client, user, API token, or resource. Each key gets its own limiting state.

## Why use one?

Without a rate limiter, a fast client can consume all the capacity of a service. A limiter protects that service and makes request admission predictable.

## Algorithms

### Fixed window

Counts requests inside fixed time blocks:

```text
window 1:  |--------- 5 seconds ---------|  count=3
window 2:  |--------- 5 seconds ---------|  counter resets
```

It is simple and uses little memory. A burst can happen at the boundary between two windows.

### Sliding-window log

Stores the timestamps of accepted requests and removes timestamps outside the rolling interval.

```text
rolling window: |---------------- 5 seconds ----------------|
timestamps:           t1       t2       t3       now
```

It gives an accurate rolling count but stores more timestamp data.

### Sliding-window counter

Stores counts for the previous and current fixed windows, then estimates the rolling count using the overlap between them.

```text
previous window                 current window
count=3                         count=1
       \\ overlap weight
        +---------------------> estimated count
```

It uses constant-size state and trades exact timestamps for an estimate.

### Token bucket

Tokens refill over time. Each accepted request spends one token.

```text
refill rate ---> [ token token token ] ---> request consumes one
                    capacity: 3
```

Capacity controls bursts. Refill rate controls the sustained request rate.

## Project structure

```text
algorithms/       fixed window, sliding log, sliding counter, token bucket
ratelimiter/      shared limiter interface and state
cmd/fixedwindow/  runnable demonstrations for all algorithms
go.mod
design.md         algorithm and concurrency details
```

## Run it

Run from the `02-rate-limiter` directory:

```powershell
go run ./cmd/fixedwindow -algorithm fixed -limit 3 -window 5s -requests 5
```

Other algorithms use the same command:

```powershell
go run ./cmd/fixedwindow -algorithm sliding -limit 3 -window 5s -requests 5
go run ./cmd/fixedwindow -algorithm sliding-counter -limit 3 -window 5s -requests 5
go run ./cmd/fixedwindow -algorithm token -capacity 3 -refill-rate 1 -requests 5
```

To see the fixed-window boundary burst:

```powershell
go run ./cmd/fixedwindow -algorithm fixed -limit 3 -window 5s -requests 3 -batches 2 -wait 5.1s
```

To compare it with a rolling window:

```powershell
go run ./cmd/fixedwindow -algorithm sliding -limit 3 -window 5s -requests 3 -batches 2 -wait 4.9s
```

## Reading the output

```text
time=14:00:00.000 batch=1 request=1 key=client-1 allowed=true
time=14:00:00.001 batch=1 request=4 key=client-1 allowed=false
```

`key` identifies the independent quota. `allowed=true` means the request passed and changed the limiter state. `allowed=false` means the algorithm rejected it because the relevant window or bucket had reached its limit.
