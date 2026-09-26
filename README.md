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
