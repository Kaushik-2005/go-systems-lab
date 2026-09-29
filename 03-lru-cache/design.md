# LRU Cache Design

## The problem

The cache keeps frequently used values in memory, but memory has a fixed capacity. When the cache is full, it removes the entry that has gone unused for the longest time.

## Why it uses a map and a list

A map answers “where is this key?” quickly. A list remembers which entry is newest and which is oldest.

```text
map[key] -> list node

head/newest <-> entry <-> entry <-> tail/oldest
```

The map gives O(1) lookup. The custom doubly linked list gives O(1) movement to the front, deletion, and eviction from the back.

`Set` creates or updates a node and moves it to the head. `Get` also moves a valid node to the head, which is why `Get` needs the write lock. `Delete` removes the node from both the list and map.

## TTL and concurrency

`SetWithTTL` stores an expiration time. `Get` removes an expired entry lazily. `CleanupExpired` scans all entries so expired data can also be removed when nobody reads its key. The background cleanup goroutine stops through a context.

One mutex protects the map, list links, values, TTL state, and size. This keeps related updates together and avoids a map/list mismatch during concurrent operations.

```text
Get, Set, Delete, eviction: O(1)
background cleanup scan:    O(n)
```

## Repository design

```text
cache/cache.go       entry map, linked list, TTL, cleanup
cmd/basic/main.go    operations and observable behavior
```

The linked list is written directly instead of using `container/list` so the LRU mechanism stays visible.
