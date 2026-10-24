# Systems

This repository contains small Go implementations of system components. Each implementation explains a mechanism that larger services use.

## Purpose

The implementations help readers understand these subjects:

- The problem that each component solves.
- The state that the component stores.
- The effects of concurrent operations.
- The behavior after a failure.
- The guarantees that the component provides.
- The costs and limits of the design.

These implementations are for study and experiments. Each mechanism is small enough for readers to inspect its state and failure behavior.

## Components

Each component has one README with descriptions, diagrams, and run commands. The section links below open the illustrated explanations.

### 1. [Load Balancer](01-load-balancer/README.md)

The load balancer selects backend servers for HTTP requests. A reverse proxy forwards each request and returns its response.

- round-robin, random, and least-connections selection.
- active health checks.
- unhealthy backend skipping and recovery.
- request timeouts and backend failure handling.
- graceful shutdown.

Algorithms and diagrams: [Load Balancer README](01-load-balancer/README.md#selection-algorithms)

### 2. [Rate Limiter](02-rate-limiter/README.md)

The rate limiter accepts or rejects requests for a key. Each key has a separate limit state.

- fixed-window counter.
- sliding-window log.
- sliding-window counter.
- token bucket.

Algorithms and diagrams: [Rate Limiter README](02-rate-limiter/README.md#the-four-algorithms)

### 3. [LRU Cache](03-lru-cache/README.md)

The cache stores values in a map and records recent use in a custom doubly linked list. It removes least recently used (LRU) entries when capacity is exceeded.

- O(1) lookup and recency updates.
- least-recently-used eviction.
- Time to live (TTL) expiration.
- concurrency-safe access.
- background cleanup.

Design and diagrams: [LRU Cache README](03-lru-cache/README.md#map-list-and-eviction)

### 4. [Consistent Hashing](04-consistent-hashing/README.md)

The hash ring assigns keys to nodes. It reduces changed assignments when nodes enter or leave the group.

- modulo baseline.
- clockwise lookup.
- node addition and removal.
- virtual nodes.
- key redistribution demonstration.

Algorithms and diagrams: [Consistent Hashing README](04-consistent-hashing/README.md#ring-lookup)

### 5. [Message Queue](05-message-queue/README.md)

The message broker stores pending work in memory. Producers and consumers can operate at different rates.

- FIFO delivery.
- concurrent producers and consumers.
- message IDs and acknowledgements.
- in-flight messages and visibility timeout.
- redelivery, retry limit, and dead-letter queue.

Design and diagrams: [Message Queue README](05-message-queue/README.md#fifo-delivery)

### 6. [Pub/Sub Broker](06-pub-sub-broker/README.md)

The pub/sub broker sends a separate event copy to each subscriber of a topic.

- multiple subscribers.
- fan-out delivery.
- subscribe and unsubscribe.
- concurrent publication.
- blocking and non-blocking slow-subscriber policies.

Design and diagrams: [Pub/Sub Broker README](06-pub-sub-broker/README.md#subscribe-and-fan-out)

### 7. [Key-Value Store](07-key-value-store/README.md)

The key-value store keeps values in memory. An optional append-only log preserves mutations for recovery.

- `Put`, `Get`, and `Delete` operations.
- JSON mutation records.
- tombstones for deletes.
- recovery by replaying the log.
- compaction of obsolete records.
- concurrency-safe access.

Design and diagrams: [Key-Value Store README](07-key-value-store/README.md#writes-and-tombstones)

### 8. [Write-Ahead Log](08-write-ahead-log/README.md)

The write-ahead log records and syncs mutations before the application changes its state.

- monotonically increasing sequence numbers.
- JSON record encoding.
- startup record validation.
- SHA-224 checksums.
- corruption detection.
- incomplete-tail recovery.

Design and diagrams: [Write-Ahead Log README](08-write-ahead-log/README.md#append-and-sync)

### 9. [Circuit Breaker](09-circuit-breaker/README.md)

The circuit breaker temporarily stops new dependency calls after repeated failures.

- closed, open, and half-open states.
- failure threshold.
- reset timeout.
- recovery probes.
- concurrency-safe state transitions.

Design and diagrams: [Circuit Breaker README](09-circuit-breaker/README.md#closed-open-and-half-open-states)

### 10. [Service Discovery](10-service-discovery/README.md)

The registry stores instance addresses under service names. Clients request current addresses from the registry.

- service registration and lookup.
- multiple instances per service.
- heartbeat and TTL leases.
- stale-instance cleanup.
- HTTP registration, lookup, heartbeat, and deregistration.

Design and diagrams: [Service Discovery README](10-service-discovery/README.md#registration-and-lookup)

### 11. [Distributed Lock](11-distributed-lock/README.md)

The lock grants one client a lease and a fencing token. The token identifies the current owner.

- acquire, renew, and release.
- lease expiration.
- owner and token validation.
- HTTP lock service.

Design and diagrams: [Distributed Lock README](11-distributed-lock/README.md#acquire-renew-and-release)

### 12. [Replicated Key-Value Store](12-replicated-kv-store/README.md)

The leader accepts writes. Followers request ordered records and apply them to their local maps.

- leader-only writes.
- ordered replication records.
- asynchronous follower updates.
- reconnect and catch-up.
- observable replication sequence.

Design and diagrams: [Replicated Key-Value Store README](12-replicated-kv-store/README.md#asynchronous-leader--follower-replication)

### 13. [Sharding Router](13-sharding-router/README.md)

The sharding router assigns keys to storage partitions. It provides hash routing and range routing.

- shard abstraction.
- deterministic key-to-shard routing.
- hash-based partitioning.
- range-based partitioning.
- per-shard in-memory map.
- shard addition and removal.
- basic key migration and rebalancing.
- key redistribution demonstration.

Design and diagrams: [Sharding Router README](13-sharding-router/README.md#hash-based-routing)

### 14. [Leader Election](14-leader-election/README.md)

The election state machine selects a coordinator through majority voting. Terms separate election attempts.

- follower, candidate, and leader states.
- term numbers.
- one vote per term.
- majority election rule.
- higher-term step-down.

Design and diagrams: [Leader Election README](14-leader-election/README.md#follower-candidate-and-leader-states)

### 15. [Distributed Cache](15-distributed-cache/README.md)

The cache model combines local caches with consistent hashing, replication, and manual health flags. All nodes exist in one process.

- custom LRU eviction.
- TTL expiration.
- concurrency-safe cache access.
- virtual-node routing and replication.
- manual node health and failover routing.
- node addition, removal, and rebalancing.

Design and diagrams: [Distributed Cache README](15-distributed-cache/README.md#local-cache-nodes)

## Component uses

A larger service can use several components. This table gives possible uses:

| Component | Purpose | Examples |
| --- | --- | --- |
| Load balancer | Sends requests to healthy servers. | Web applications, API gateways, ingress controllers. |
| Rate limiter | Limits requests for a client or resource. | Login protection, public APIs, payment and messaging quotas. |
| LRU cache | Keeps recent values in memory. | Database query caching, session data, computed results. |
| Consistent hashing | Assigns keys with limited changes during node addition or removal. | Distributed caches, sharded databases, partitioned message systems. |
| Message queue | Stores work until a consumer can process it. | Email sending, image processing, background jobs, order fulfillment. |
| Pub/Sub broker | Sends one event independently to many interested subscribers. | Notifications, analytics events, audit streams, service integration. |
| Key-value store | Stores and retrieves data by a simple key. | Configuration, feature flags, metadata, local embedded storage. |
| Write-ahead log | Records mutations before application state changes. | Databases, replicated logs, durable queues, storage engines. |
| Circuit breaker | Stops new calls after repeated dependency failures. | Service-to-service calls, payment providers, external APIs. |
| Service discovery | Helps clients find the current instances of a service. | Microservices, container platforms, dynamically scaled workers. |
| Distributed lock | Coordinates work so only one client owns a shared operation. | Leader selection, scheduled jobs, migrations, duplicate-work prevention. |
| Replicated key-value store | Copies ordered changes from a leader to followers. | Primary/replica databases, configuration stores, read scaling. |
| Sharding router | Decides which storage partition owns a key. | Distributed databases, partitioned caches, large-scale key-value services. |
| Leader election | Chooses one coordinator among several nodes. | Replicated databases, cluster controllers, distributed schedulers. |
| Distributed cache | Stores values across several cache nodes. | Database caching, session lookup, computed results, API response caching. |

An order service can use a load balancer and rate limiter for incoming requests. A circuit breaker can protect calls to its payment provider. A queue can store fulfillment work. A replicated store can retain copies of order data.

## Development process

Use this process to study a component:

```text
Read the mechanism.
      |
      v
Run the example.
      |
      v
Examine the state and failure behavior.
      |
      v
Change the mechanism.
      |
      v
Compare the results.
```

Each README explains data structures, algorithms, concurrency, and failure behavior. It also describes implementation limits and gives run commands.

## Documentation

Use [ASD-STE100 Simplified Technical English](https://www.asd-ste100.org/) for documentation updates.

- Describe one topic in each sentence.
- Use the active voice.
- Give one instruction in each procedural sentence.
- Limit procedural sentences to 20 words.
- Limit descriptive sentences to 25 words.
- Use one technical term for each concept.
- Define an abbreviation before its first use.
- Preserve API identifiers, command syntax, and literal output.

Use standard software terms as technical nouns or technical verbs. Explain their meaning when the context does not define them.

## Language and dependencies

The implementations use Go and prefer the standard library. Each component implements the mechanism directly instead of delegating it to an external service.
