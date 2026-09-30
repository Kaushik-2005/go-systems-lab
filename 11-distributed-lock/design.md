# Distributed Lock Design

## The idea

Several clients may want to update the same resource. The lock gives one client ownership for a limited time. The time limit matters because a process can crash while holding the lock.

This is an in-memory model first, followed by a small HTTP server so the same mechanism can be used by separate local processes.

## State and flow

The lock stores:

- whether it is held;
- the current owner ID;
- the current fencing token;
- the next token to issue;
- the lease expiration time.

`Acquire` takes the mutex, clears an expired lease, and either rejects the request or increments `nextToken` and creates a new lease. The counter is never reset, so every successful acquisition gets a larger token.

`Renew` and `Release` require both the owner ID and the current token. An old owner cannot accidentally modify a newer lease after its own lease expires.

```text
acquire(owner)
       |
       v
 free? ── no ──> rejected
   |
  yes
   |
   v
 issue token + expiration
       |
       v
 renew before expiry, or release
       |
       v
 expiry makes the lock available
```

## Concurrency and HTTP

The `Lock` uses one `sync.Mutex` because every operation reads and updates the same small piece of state. Expiration is checked while holding that mutex, so two concurrent acquisitions cannot both win.

The HTTP server keeps one `Lock` instance and exposes `POST /acquire`, `POST /renew`, and `POST /release`. HTTP handlers can run concurrently, and the lock serializes their state changes.

## Why fencing tokens matter

A lease tells the lock service who currently owns the lock, but an old client may wake up after its lease expires and continue its work. The new owner receives a larger token. If a protected resource remembers the newest token and rejects smaller ones, delayed work from the old owner cannot overwrite newer work.

This repository generates the tokens and checks them at the lock boundary. It does not include a separate protected resource that enforces them.

## Failure behavior

If an owner disappears, its lease eventually expires. If it renews too slowly, another client may acquire the lock. A release or renewal from the old owner is rejected because its owner/token pair is no longer current.

The state is only in memory. Restarting the service loses the lock and token history, and running multiple server processes does not create a coordinated distributed lock. This is a focused model of the lease and fencing mechanisms, not a consensus-backed production lock service.
