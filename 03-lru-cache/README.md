# LRU Cache

## What it is

Lets say reading from a database is slow but the same data is requested often. A cache keeps recently used values in memory so the next read is faster.

```text
request -> cache -> hit: return value
              -> miss: load and store value
```

## How it works

The cache stores entries in most-recently-used to least-recently-used order.

```text
newest -> [entry] <-> [entry] <-> [oldest]
```

When the cache is full, it removes the least recently used entry. A map provides fast key lookup and a custom doubly linked list keeps the usage order.

It also handles:

- `Get`, `Set`, and `Delete`;
- fixed capacity;
- TTL expiration;
- background cleanup;
- concurrent access.

## Project structure

```text
cache/       map, linked list, TTL, and cleanup
cmd/basic/   cache demonstration
go.mod
```

## Run it

Run from the `03-lru-cache` directory:

```powershell
go run ./cmd/basic
```

The demo shows a read changing recency order, an old entry being evicted, TTL expiration, concurrent access, and background cleanup.
