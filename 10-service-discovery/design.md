# Service Discovery Design

## The problem

Service addresses can change when processes restart, ports change, or instances are added and removed. A registry gives clients a name-based way to find live instances.

```text
register -> registry -> lookup
heartbeat -> renew lease
timeout -> remove stale instance
```

## Registry and leases

The registry stores instances by service name and instance ID. Each entry contains an address and expiration time. `Register` creates or replaces an entry, `Lookup` returns live entries, and `Deregister` removes one explicitly.

`Heartbeat` extends the expiration time. If a service stops heartbeating, the background cleanup goroutine removes it after the TTL. Lookup also removes expired entries before returning results.

## HTTP API

The server exposes:

```text
POST   /register
GET    /lookup?service=payments
GET    /heartbeat?service=payments&id=payments-1
DELETE /deregister?service=payments&id=payments-1
```

The HTTP layer validates query parameters and delegates state changes to the registry package. The registry itself does not depend on HTTP.

## Concurrency and repository design

The registry mutex protects service membership, instance metadata, and expiration times. The cleanup goroutine uses a context for shutdown and a ticker for periodic removal.

```text
registry/registry.go  instances, TTLs, heartbeats, cleanup
cmd/demo/main.go      in-process behavior
cmd/server/main.go    HTTP API
```

The registry is in memory and local to one process. It does not persist registrations, replicate state, or provide consensus between registry servers.
