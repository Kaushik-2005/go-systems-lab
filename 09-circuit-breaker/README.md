# Circuit Breaker

## What it is

Lets say there is an API you are calling and it is failing due to a server side issue. You don't want to keep sending request to that, so a Circuit Breaker will help you in stopping the requests to that API.

```text
caller -> circuit breaker -> dependency
```

## States

```text
CLOSED -> OPEN -> HALF_OPEN -> CLOSED
                   |
                   +---------> OPEN
```

- `closed`: calls reach the dependency;
- `open`: calls are rejected immediately;
- `half-open`: one probe tests whether the dependency recovered.

## Project structure

```text
breaker/breaker.go       state machine and synchronized execution
cmd/demo/main.go         failure and recovery demonstration
cmd/concurrent/main.go   single-probe concurrency demonstration
go.mod
```

## Run it

```powershell
go run .\cmd\demo\main.go
go run .\cmd\concurrent\main.go
```

The demos show failure thresholds, reset timeouts, failed and successful recovery probes, and concurrent callers competing for one half-open probe.
