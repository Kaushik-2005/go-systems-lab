# Circuit Breaker

## What it is

A circuit breaker wraps calls to a dependency and temporarily stops admitting new calls after repeated failures. This avoids spending resources on a dependency that is likely to fail again.

## How it works

### Closed, open, and half-open states

![The breaker opens after a failure threshold, admits a recovery probe after a timeout, and closes on probe success](diagrams/state-transitions.svg)

`New(threshold, resetTimeout)` starts in `closed`. `Execute(operation)` checks admission, runs the supplied function if admitted, and records its result.

| State | New call behavior | Transition |
| --- | --- | --- |
| Closed | Run the operation | Reaching the failure threshold opens the circuit |
| Open | Return `ErrOpen` before the timeout | The next call after the timeout enters half-open |
| Half-open | Admit one recovery probe | Success closes; failure reopens and restarts the timeout |

An error increments the failure counter. A success resets it to zero and sets the state to closed. In sequential use, the threshold therefore counts consecutive failures. An admitted operation returns its own error; a rejected call returns `ErrOpen` without invoking the dependency.

The timeout is checked lazily on `Execute`. Waiting alone does not change the stored state returned by `State()`.

### One recovery probe

![One half-open caller reserves the probe flag; other callers are rejected while it runs](diagrams/half-open-probe.svg)

When the open timeout has elapsed, the first caller sets `probeRunning` under the mutex. Other half-open callers see the flag and return `ErrOpen`. The admitted function runs outside the mutex, allowing other callers to check the state while the dependency is slow.

A failed probe records a new `openedAt` and returns to open. A successful probe resets failures and returns to closed. The diagram assumes earlier closed-state operations have already finished.

## Concurrency and limits

The mutex protects state, counters, timestamps, and probe admission. It does not cancel calls that were admitted earlier. Those calls still update the shared state when they complete; there is no generation number to distinguish their results from a later recovery cycle. A late success can therefore close a breaker that another call has opened.

This is an in-process state machine. It does not provide request retries, operation timeouts, panic recovery, durable state, or coordination across service processes. A successful probe is one observation of the dependency, rather than a guarantee that later calls will succeed.

## Project structure

```text
breaker/breaker.go       state machine and synchronized execution
cmd/demo/main.go         sequential failure and recovery demo
cmd/concurrent/main.go   concurrent recovery-probe demo
diagrams/               SVG illustrations
go.mod
```

## Run it

Run from the `09-circuit-breaker` directory:

```powershell
go run ./cmd/demo
go run ./cmd/concurrent
```

The demos show failure thresholds, open-call rejection, reset timeouts, failed and successful recovery probes, and concurrent callers competing for the half-open probe.
