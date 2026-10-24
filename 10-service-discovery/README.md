# Service Discovery

## What it is

A service registry stores instance addresses under a service name. Clients use that name to find an address. They do not need one fixed machine or port.

## How it works

### Registration and lookup

![An instance registers its address. A client reads that address and connects directly to the instance.](diagrams/registration-and-lookup.svg)

The registry maps each service name to instance registrations. Each registration has an instance ID, address, and expiration time. `Register` creates or replaces the entry for that service and ID. It sets `expiresAt = now + TTL`. TTL means time to live.

Multiple instances can use the same service name. `Lookup(service)` removes expired registrations before it returns the remaining instances. The result order is unspecified because the registry uses a map.

The client selects an address and connects directly to it. The registry does not forward application traffic. It does not distribute requests across the returned addresses.

### Heartbeats and expiration

![A heartbeat extends a lease. A later expiration check removes the instance.](diagrams/heartbeat-and-expiration.svg)

`Heartbeat(service, id, TTL)` extends a known, unexpired registration. It sets expiration relative to the current time. If the entry is missing or expired, the function returns `false`. The instance must register again.

`StartCleanup(ctx, interval)` checks all registrations on a ticker. It stops when the context is canceled. Lookup also checks expiration between background checks. `Deregister` removes an instance explicitly. If the service has no instances, it removes the service map too.

A lease shows recent contact with the registry. It does not prove that the application endpoint works. An instance can fail immediately after a heartbeat or lookup. A returned address can therefore be unavailable.

### HTTP API

The command listens on `:8080`. It uses 30-second registration and heartbeat leases. It checks expiration every 100 ms. It also creates a `payments-1` registration for `localhost:9001` at startup.

| Request | Effect |
| --- | --- |
| `POST /register?service=...&id=...&address=...` | Create or replace a registration |
| `GET /lookup?service=...` | Return current instances |
| `GET /heartbeat?service=...&id=...` | Extend an unexpired registration |
| `DELETE /deregister?service=...&id=...` | Remove an instance |

The HTTP handlers check required query parameters. They call the registry package for state changes. Heartbeat returns `404` for a missing or expired instance. Deregister returns `404` for a missing instance.

## Concurrency, cost, and limits

One mutex protects membership and expiration state. Lookup uses the write lock because it removes expired entries. Register, heartbeat, and deregister take O(1) time on average. Cleanup and lookup check all registrations for expiration. Their cost increases with the total instance count.

All registrations exist in one process. A restart loses them. Different registry processes do not share state or use consensus. The registry package does not depend on HTTP.

## Project structure

```text
registry/       instance state, leases, heartbeats, and cleanup
cmd/demo/       in-process registration demonstration
cmd/server/     HTTP registry server
diagrams/       SVG illustrations
go.mod
```

## Run it

Open a terminal in the `10-service-discovery` directory.
Run this command:

```powershell
go run .\cmd\server\main.go
```

Register an instance:

```powershell
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/register?service=payments&id=payments-2&address=localhost:9002"
```

Read the registered instances:

```powershell
Invoke-RestMethod "http://localhost:8080/lookup?service=payments"
```

Renew its lease:

```powershell
Invoke-RestMethod "http://localhost:8080/heartbeat?service=payments&id=payments-2"
```

Deregister it:

```powershell
Invoke-RestMethod -Method Delete -Uri "http://localhost:8080/deregister?service=payments&id=payments-2"
```
