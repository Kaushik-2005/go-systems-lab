# Distributed Lock

## What it is

A lock service lets clients coordinate access to one shared operation. This component models one in-memory lock with a lease: ownership expires if the holder stops renewing it.

## How it works

### Acquire, renew, and release

![One client acquires and renews a lease; a competing client is rejected until expiration permits a new owner](diagrams/lease-lifecycle.svg)

The lock stores whether it is held, the current owner, current token, next token counter, and expiration time. `Acquire(owner, TTL)` clears an expired lease, rejects acquisition if the lock is still held, or increments the token counter and grants a new lease.

`Renew(owner, token, TTL)` extends a valid current lease to `now + TTL` without changing its token. `Release(owner, token)` clears current ownership. Both reject a stale owner/token pair or an expired lease. Even the current owner cannot call acquire again while its lease is held; renewal is a separate operation.

All three operations check expiration while holding one mutex. Expiration is lazy: a background reaper is unnecessary because the next operation detects an expired holder. The mutex ensures concurrent acquisitions cannot both win within this lock instance.

### Fencing tokens

![A newer owner receives a larger token; old-token renewals and releases fail, while downstream writes need separate fencing enforcement](diagrams/fencing-tokens.svg)

A can pause while holding token 41, lose its lease, and later resume after B has acquired token 42. A's old owner/token pair cannot release or renew B's lock.

Preventing A from modifying the shared resource requires an additional check at that resource. It must remember the latest accepted fencing token and reject older ones. The repository issues tokens and validates them at the lock boundary; it does not implement a downstream resource that enforces fencing on its own writes.

### HTTP service

The server holds one `Lock` instance, so concurrent HTTP handlers share the same ownership state. The examples use these requests:

| Endpoint | Parameters | Response fields |
| --- | --- | --- |
| `POST /acquire` | `owner`, positive `ttl_ms` | `acquired`, `token` |
| `POST /renew` | `owner`, `token`, positive `ttl_ms` | `renewed` |
| `POST /release` | `owner`, `token` | `released` |

## Concurrency and limits

Each ownership operation is O(1) and serialized by a `sync.Mutex`. An expired lease frees the lock for a new caller but does not stop the old process from running.

The counter increases only within a server lifetime. Restart loses both ownership and token history, and independent server processes do not coordinate. This is a model of leases and fencing, without replicated state or consensus-backed lock ownership.

## Project structure

```text
lock/lock.go        lease and fencing-token mechanism
cmd/demo/main.go    in-process demonstration
cmd/server/main.go  HTTP lock service
diagrams/           SVG illustrations
go.mod
```

## Run it

Run from the `11-distributed-lock` directory. Start the in-process demo:

```bash
go run ./cmd/demo
```

Run the HTTP service:

```bash
go run ./cmd/server -addr :8080
```

Use another port if `8080` is already in use:

```bash
go run ./cmd/server -addr :8090
```

Acquire, renew, and release with PowerShell:

```powershell
$lock = Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8090/acquire?owner=client-a&ttl_ms=5000"

Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8090/renew?owner=client-a&token=$($lock.token)&ttl_ms=5000"

Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8090/release?owner=client-a&token=$($lock.token)"
```

The service is an educational single-process lock server. It demonstrates leases and fencing tokens, but it does not replicate its state or provide consensus between multiple lock servers.
