# LRU Cache Design

## What it solves

The cache stores recently used key-value pairs in memory. A repeated lookup can return quickly without repeating slower work.

The cache has a fixed capacity. When it is full, it removes the entry that has been unused for the longest time.

## The map-plus-list design

Two data structures are needed because they answer different questions:

```text
map[string]*entry
        |
        +---- "language" ----> list node

linked list:
head <-> newest <-> ... <-> oldest <-> tail
```

The map answers “where is this key?” in O(1). The doubly linked list answers “which entry is newest or oldest?” and moves existing nodes in O(1).

The list rules are:

- `head` is the most recently used entry;
- `tail` is the least recently used entry;
- every `prev` link and `next` link agrees with its neighbor;
- the map contains exactly the entries in the list.

## Operations

### `Set(key, value)`

If the key exists, the cache updates its value and moves the node to the head. If it is new, the cache creates a node, puts it at the head, and stores it in the map. When the size exceeds capacity, the tail node is removed from both structures.

### `Get(key)`

The cache looks up the node in the map. A missing or expired node produces a miss. A valid node moves to the head before its value is returned.

`Get` needs the write lock because a read changes recency order.

### `Delete(key)`

The cache finds the node, unlinks it from the list, and deletes it from the map.

## TTL expiration

`SetWithTTL` adds an expiration time to an entry:

```text
SetWithTTL("session", "active", 500ms)
             |
          valid
             |
        500ms passes
             |
          expired
```

Normal `Get` calls remove expired entries lazily. `CleanupExpired` scans the cache so an expired entry can be removed even when nobody reads its key.

## Concurrency model

The cache uses one mutex to protect the map, list links, TTL state, and size. `Set`, `Get`, `Delete`, and cleanup all update related pieces of state atomically from the caller's perspective.

`StartCleanup(ctx, interval)` starts a ticker-based goroutine. Cancelling the context stops the goroutine and releases its ticker.

## Complexity

```text
Get       O(1)
Set       O(1)
Delete    O(1)
move node O(1)
eviction  O(1)
```

Lazy expiration for a known key is O(1). A full cleanup scan is O(n), where n is the number of entries.

## Repository design

```text
cache/cache.go       entry map, doubly linked list, TTL, cleanup
cmd/basic/main.go    operations and observable demonstration
```

The custom linked list is implemented directly instead of using `container/list`, so the map-plus-list mechanism stays visible.
