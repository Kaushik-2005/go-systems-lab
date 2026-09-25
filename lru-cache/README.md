# LRU Cache

## What is a cache?

A cache stores recently used data closer to the caller so repeated reads can avoid slower work such as computation, disk access, or a network request.

When a cache has limited capacity, it needs an eviction policy. Least Recently Used (LRU) removes the entry that has gone unused for the longest time.

```text
request -> cache lookup -> hit: return value
                     |
                     `-> miss: load value, store it, return it
```

This component will build an in-memory LRU cache from scratch in Go.

## Learning sequence

1. Basic `Get`, `Set`, and `Delete` operations.
2. Fixed capacity and LRU eviction.
3. O(1) lookup and recency updates.
4. TTL expiration.
5. Concurrency safety and optional cleanup.

## Project structure

```text
cache/            cache state and operations
cmd/              runnable demonstrations
go.mod            Go module definition
design.md         design reasoning and trade-offs
```

## Current status

The component is at the core-operation design checkpoint. No implementation has been added yet.
