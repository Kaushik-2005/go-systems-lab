# Systems

This repository is my hands-on path to understanding system design and distributed systems by building the mechanisms behind common infrastructure components in Go.

## Why I am building this

I am building simplified systems from scratch so I can understand:

- what problem each component solves;
- what state it needs to maintain;
- how concurrent operations interact;
- what happens when something fails;
- which guarantees the component provides;
- what trade-offs shape the design.

The goal is not production infrastructure. The goal is to make each mechanism small enough to read, run, break, and understand.

## Components

### 1. [Load Balancer](01-load-balancer/README.md)

An HTTP reverse proxy that distributes requests across backend servers.

- round-robin, random, and least-connections selection;
- active health checks;
- unhealthy backend skipping and recovery;
- request timeouts and backend failure handling;
- graceful shutdown.

Design: [01-load-balancer/design.md](01-load-balancer/design.md)

### 2. [Rate Limiter](02-rate-limiter/README.md)

Request-admission algorithms that decide whether a keyed request is accepted.

- fixed-window counter;
- sliding-window log;
- sliding-window counter;
- token bucket.

Design: [02-rate-limiter/design.md](02-rate-limiter/design.md)

### 3. [LRU Cache](03-lru-cache/README.md)

An in-memory cache using a map and custom doubly linked list.

- O(1) lookup and recency updates;
- least-recently-used eviction;
- TTL expiration;
- concurrency-safe access;
- background cleanup.

Design: [03-lru-cache/design.md](03-lru-cache/design.md)

### 4. [Consistent Hashing](04-consistent-hashing/README.md)

A hash ring that routes keys to nodes while minimizing movement when nodes change.

- modulo baseline;
- clockwise lookup;
- node addition and removal;
- virtual nodes;
- key redistribution demonstration.

Design: [04-consistent-hashing/design.md](04-consistent-hashing/design.md)

### 5. [Message Queue](05-message-queue/README.md)

An in-memory broker that separates producers from consumers.

- FIFO delivery;
- concurrent producers and consumers;
- message IDs and acknowledgements;
- in-flight messages and visibility timeout;
- redelivery, retry limit, and dead-letter queue.

Design: [05-message-queue/design.md](05-message-queue/design.md)

### 6. [Pub/Sub Broker](06-pub-sub-broker/README.md)

A topic-based broker that sends each event to every subscriber of that topic.

- multiple subscribers;
- fan-out delivery;
- subscribe and unsubscribe;
- concurrent publication;
- blocking and non-blocking slow-subscriber policies.

Design: [06-pub-sub-broker/design.md](06-pub-sub-broker/design.md)

### 7. [Key-Value Store](07-key-value-store/README.md)

An in-memory key-value store backed by an append-only log.

- `Put`, `Get`, and `Delete` operations;
- JSON mutation records;
- tombstones for deletes;
- recovery by replaying the log;
- compaction of obsolete records;
- concurrency-safe access.

Design: [07-key-value-store/design.md](07-key-value-store/design.md)

### 8. [Write-Ahead Log](08-write-ahead-log/README.md)

An append-only durability log that records ordered mutations before state is applied.

- monotonically increasing sequence numbers;
- JSON record encoding;
- replay during startup;
- SHA-256 checksums;
- corruption detection;
- incomplete-tail recovery.

Design: [08-write-ahead-log/design.md](08-write-ahead-log/design.md)

### 9. [Circuit Breaker](09-circuit-breaker/README.md)

A fault-tolerance wrapper that stops calling a repeatedly failing dependency.

- closed, open, and half-open states;
- failure threshold;
- reset timeout;
- recovery probes;
- concurrency-safe state transitions.

Design: [09-circuit-breaker/design.md](09-circuit-breaker/design.md)

### 10. [Service Discovery](10-service-discovery/README.md)

An in-memory registry where services register addresses and clients discover live instances.

- service registration and lookup;
- multiple instances per service;
- heartbeat and TTL leases;
- stale-instance cleanup;
- HTTP registration, lookup, heartbeat, and deregistration.

Design: [10-service-discovery/design.md](10-service-discovery/design.md)

### 11. [Distributed Lock](11-distributed-lock/README.md)

A lease-based lock that gives one client ownership at a time and returns fencing tokens to reject stale clients.

- acquire, renew, and release;
- lease expiration;
- owner and token validation;
- HTTP lock service.

Design: [11-distributed-lock/design.md](11-distributed-lock/design.md)

### 12. [Replicated Key-Value Store](12-replicated-kv-store/README.md)

A leader/follower store where the leader accepts writes and followers pull ordered records to keep local copies.

- leader-only writes;
- ordered replication records;
- asynchronous follower updates;
- reconnect and catch-up;
- observable replication sequence.

Design: [12-replicated-kv-store/design.md](12-replicated-kv-store/design.md)

### 13. [Sharding Router](13-sharding-router/README.md)

A routing layer that maps keys to storage shards using deterministic hashing.

- shard abstraction;
- deterministic key-to-shard routing;
- hash-based partitioning;
- range-based partitioning;
- per-shard in-memory map;
- shard addition and removal;
- basic key migration and rebalancing;
- key redistribution demonstration.

Design: [13-sharding-router/design.md](13-sharding-router/design.md)

### 14. [Leader Election](14-leader-election/README.md)

A state machine that chooses one coordinator from a group of nodes using terms and majority voting.

- follower, candidate, and leader states;
- term numbers;
- one vote per term;
- majority election rule;
- higher-term step-down.

Design: [14-leader-election/design.md](14-leader-election/design.md)

### 15. [Distributed Cache](15-distributed-cache/README.md)

A cache node that will be composed into a distributed cache using routing, replication, and health awareness.

- custom LRU eviction;
- TTL expiration;
- concurrency-safe cache access.

Design: [15-distributed-cache/design.md](15-distributed-cache/design.md)

## Where these components are used

These components usually appear together inside larger services. Each one solves a different problem:

| Component | Purpose | Real-world examples |
| --- | --- | --- |
| Load balancer | Spreads incoming traffic across healthy servers. | Web applications, API gateways, ingress controllers. |
| Rate limiter | Controls how often a client or resource can be used. | Login protection, public APIs, payment and messaging quotas. |
| LRU cache | Keeps frequently reused data close to the application. | Database query caching, session data, computed results. |
| Consistent hashing | Routes a key to a node while reducing movement when nodes change. | Distributed caches, sharded databases, partitioned message systems. |
| Message queue | Stores work until a consumer can process it. | Email sending, image processing, background jobs, order fulfillment. |
| Pub/Sub broker | Sends one event independently to many interested subscribers. | Notifications, analytics events, audit streams, service integration. |
| Key-value store | Stores and retrieves data by a simple key. | Configuration, feature flags, metadata, local embedded storage. |
| Write-ahead log | Records mutations so state can be recovered after a crash. | Databases, replicated logs, durable queues, storage engines. |
| Circuit breaker | Stops calling a failing dependency so failures do not spread. | Service-to-service calls, payment providers, external APIs. |
| Service discovery | Helps clients find the current instances of a service. | Microservices, container platforms, dynamically scaled workers. |
| Distributed lock | Coordinates work so only one client owns a shared operation. | Leader selection, scheduled jobs, migrations, duplicate-work prevention. |
| Replicated key-value store | Keeps copies of data across processes for availability and reads. | Primary/replica databases, configuration stores, read scaling. |
| Sharding router | Decides which storage partition owns a key. | Distributed databases, partitioned caches, large-scale key-value services. |
| Leader election | Chooses one coordinator among several nodes. | Replicated databases, cluster controllers, distributed schedulers. |
| Distributed cache | Keeps frequently used data close to applications across several nodes. | Database caching, session lookup, computed results, API response caching. |

For example, an order service might use a load balancer to receive traffic, a rate limiter to protect its API, a circuit breaker around the payment provider, a message queue for fulfillment work, and a replicated store for order data.

## Building approach

Each component is developed through observable steps:

```text
small mechanism
      |
      v
run it and observe behavior
      |
      v
find a failure or trade-off
      |
      v
add the concept that addresses it
```

Every component has runnable Go code, a practical README, and a design document explaining its data structures, algorithms, concurrency, and failure behavior.

## Language and dependencies

Implementations are written in Go and prefer the standard library. External systems are not used to replace the mechanism being studied.
