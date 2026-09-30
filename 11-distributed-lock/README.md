# Distributed Lock

## What it is

A distributed lock lets multiple clients coordinate access to one shared resource. Only one client should hold the lock at a time.

This component models a lock service with leases. A lease expires if the owner stops renewing it, so a crashed client does not hold the lock forever.

```text
client A ── acquire ──┐
                      v
                 lock service
                      ^
client B ── acquire ──┘
                 only one wins
```

## How it works

- `Acquire` returns a unique, increasing fencing token when the lock is free.
- The owner must periodically call `Renew` before the lease expires.
- `Release` only works with the current owner and its token.
- An expired lease can be acquired by another client.
- The fencing token changes on every successful acquisition. A stale client with an old token cannot release or renew the newer lock.

The token is useful to a downstream resource too: that resource can reject an operation carrying a token older than the latest token it has accepted.

## Project structure

```text
11-distributed-lock/
├── lock/lock.go          # lease and fencing-token mechanism
└── cmd/
    ├── demo/main.go      # in-process demonstration
    └── server/main.go    # HTTP lock service
```

## Run it

Run the in-process demo:

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
