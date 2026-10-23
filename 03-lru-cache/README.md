# LRU Cache

## What it is

Let's say reading from a database is slow but the same data is requested often. A cache keeps recently used values in memory so the next read is faster.

```text
request -> cache -> hit: return value
              -> miss: load and store value
```

## How it works

The cache stores entries in most-recently-used to least-recently-used order.

When the cache is full, it removes the least recently used entry. A map provides fast key lookup and a custom doubly linked list keeps the usage order.

It also handles:

- `Get`, `Set`, and `Delete`;
- fixed capacity;
- TTL expiration;
- background cleanup;
- concurrent access.

### Map, list, and eviction

![A map finds list entries; Get moves A to the head and inserting D evicts the least recently used entry B](diagrams/recency-and-eviction.svg)

The map points directly to list entries. The head is most recently used and the tail is least recently used. A map lookup finds an entry in O(1), and its `prev` and `next` pointers allow O(1) movement or removal.

`Set` creates or updates an entry and moves it to the head. `Get` also moves a valid entry to the head. If insertion exceeds capacity, the tail is removed from both the list and map. `Delete` removes the requested entry from both structures.

The linked list is written directly instead of using `container/list` so the LRU mechanism stays visible.

### TTL expiration and cleanup

![A five-second TTL allows reads before expiration and removes expired entries on reads or during cleanup](diagrams/ttl-and-cleanup.svg)

`SetWithTTL` stores an expiration time. `Get` removes an expired entry lazily and returns a miss. `CleanupExpired` scans all entries so expired data can also be removed when nobody reads its key. `StartCleanup` runs those scans on a ticker and stops when its context is canceled. Plain `Set` stores a value without an expiration time.

`Len` and `Keys` can include expired entries until a read or cleanup removes them. `Keys` returns entries from most recently used to least recently used.

## Concurrency and cost

One mutex protects the map, list links, values, TTL state, and size. This keeps related updates together and avoids a map/list mismatch during concurrent operations. `Get` takes the same exclusive lock as writes because it updates recency and can remove an expired entry.

| Operation | Time complexity |
| --- | --- |
| `Get`, `Set`, `SetWithTTL`, `Delete`, eviction | O(1) |
| `Len` | O(1) |
| `Keys`, `CleanupExpired` | O(n) |

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
