# Service Discovery

## What it is

Lets say a service can run on different machines or ports. Instead of hardcoding its address, a client asks a registry where the service instances are.

```text
service -> registry -> client lookup
```

## How it works

Services register an ID and address under a service name. The registration has a TTL, so the service must send heartbeats to remain discoverable.

The registry also handles:

- multiple instances per service;
- lookup by service name;
- HTTP registration;
- HTTP heartbeats;
- TTL expiration and background cleanup;
- HTTP deregistration.

## Project structure

```text
registry/       service state, leases, heartbeats, and cleanup
cmd/demo/       in-process registration demonstration
cmd/server/     HTTP registry server
go.mod
```

## Run it

Run from the `10-service-discovery` directory:

```powershell
go run .\cmd\server\main.go
```

Register an instance:

```powershell
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/register?service=payments&id=payments-2&address=localhost:9002"
```

Look it up:

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
