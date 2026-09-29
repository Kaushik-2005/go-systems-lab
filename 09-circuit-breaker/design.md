# Circuit Breaker Design

## The problem

Repeated calls to a failing dependency waste resources and can spread the failure through the rest of the system. The breaker temporarily stops those calls.

```text
caller -> circuit breaker -> dependency
```

## State machine

```text
CLOSED -> OPEN -> HALF_OPEN -> CLOSED
                   |
                   +---------> OPEN
```

`CLOSED` allows calls and counts failures. When the threshold is reached, the breaker records the time and becomes `OPEN`. Open calls return `ErrOpen` without reaching the dependency.

After the reset timeout, one caller enters `HALF_OPEN` as a recovery probe. A successful probe closes the circuit. A failed probe opens it again and starts another timeout.

## Concurrency and repository design

One mutex protects the state, failure count, timestamps, and probe flag. The dependency function runs outside the mutex, so a slow dependency does not block state checks. The probe flag makes sure concurrent callers do not send multiple recovery requests.

```text
breaker/breaker.go       state machine and execution wrapper
cmd/demo/main.go         sequential failure and recovery demo
cmd/concurrent/main.go   concurrent half-open callers
```

The breaker is an in-process wrapper. It does not retry requests, persist state, coordinate between processes, or prove that the dependency is healthy beyond its probe result.
