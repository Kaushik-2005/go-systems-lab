# Systems

This repository is my hands-on path to understanding system design and distributed systems by building the mechanisms behind common infrastructure components.

## Why I am building this

I do not know System Design or Distributed Systems well. Whenever I see few system design questions on interviews, I feel unprepared. Whenever I read a system design book/blog, I feel like I am missing the intuition behind the concepts. Whenever I read a distributed systems paper, I feel like I am missing the context behind the algorithms.

So, I am building simplified versions of these components from scratch so I can understand:

- what problem each component solves;
- what state it needs to maintain;
- how concurrent operations interact;
- what happens when something fails;
- which guarantees the component actually provides;
- and what trade-offs make one design different from another.

The goal is not to create production infrastructure. The goal is to make each mechanism small enough to read, run, break, and understand.

## What I am learning

The components explore several layers of systems work:

```text
traffic and concurrency
        |
        v
partitioning and communication
        |
        v
storage internals
        |
        v
resilience and membership
        |
        v
distributed coordination and data
        |
        v
composition
```

Together, they cover:

- Go concurrency and synchronization;
- HTTP and network communication;
- request admission and traffic distribution;
- caching and partitioning;
- queues and publish/subscribe messaging;
- append-only storage and recovery;
- leases, health, and failure detection;
- replication, sharding, and leader election.

## Building approach

Each component is developed through observable steps:

```text
small mechanism
      |
      v
run it and observe behavior
      |
      v
find a limitation or failure
      |
      v
add the concept that addresses it
```

For example, the Load Balancer began as a single reverse proxy, then grew to support backend selection, health checks, request timeouts, failure handling, and graceful shutdown. The Rate Limiter began with a fixed-window counter, then exposed its boundary burst before adding sliding windows and token buckets.

## Components

### Load Balancer

An HTTP load balancer with:

- round-robin, random, and least-connections selection;
- active and concurrent health checks;
- unhealthy-backend skipping and recovery;
- request timeouts and backend failure handling;
- graceful shutdown.

### Rate Limiter

Per-key request admission algorithms:

- fixed-window counter;
- sliding-window log;
- sliding-window counter;
- token bucket.

### Further components

The repository will continue with:

- LRU cache;
- consistent hashing;
- message queue;
- pub/sub broker;
- persistent key-value store;
- write-ahead log;
- circuit breaker;
- service discovery;
- distributed lock;
- replicated key-value store;
- sharding router;
- leader election;
- distributed cache.

## Language and dependencies

Implementations are written in Go and prefer the standard library. External systems are not used to replace the mechanism being studied.

Each component is intentionally small, runnable locally, and documented in terms of its data structures, algorithms, concurrency model, failure behavior, and limitations.
