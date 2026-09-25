# LRU Cache Design

## 1. What the cache must do

The cache stores key-value pairs in memory and serves repeated reads without repeating slower work.

```text
Get(key)
   |
   +--> key exists and is valid: return value
   |
   `--> key missing or expired: return cache miss
```

The cache has a fixed capacity. When it is full, it must remove one entry before accepting another. LRU chooses the entry that has been unused for the longest time.

## 2. The map-plus-list design

The cache uses two structures because each solves a different problem:

```text
map[string]*entry
        |
        +---- "language" ----> entry node
                              /          \
                         previous       next

linked list:
front <-> newest <-> ... <-> oldest <-> back
```

The map answers “where is this key?” in O(1). The doubly linked list answers “which entry is newest or oldest?” and allows an existing node to move in O(1).

The list invariant is:

- `head` is the most recently used entry;
- `tail` is the least recently used entry;
- every node's `prev` and `next` links agree;
- the map contains exactly the nodes in the list.

## 3. Operation behavior

### `Set(key, value)`

- if the key exists, update its value and move it to the front;
- otherwise create a node, add it to the front, and store it in the map;
- if the map exceeds capacity, remove the tail node from both structures.

### `Get(key)`

- look up the node in the map;
- return a miss if it does not exist;
- return a miss and remove it if its TTL has expired;
- otherwise move it to the front and return its value.

Although `Get` looks like a read, it changes recency order. It therefore needs the exclusive mutex.

### `Delete(key)`

- look up the node;
- unlink it from the list;
- delete it from the map.

## 4. TTL behavior

An entry may have no expiry or an expiry time set by `SetWithTTL`.

```text
SetWithTTL("session", "active", 500ms)
          |
          v
      valid entry
          |
       500ms passes
          |
          v
      Get -> remove entry -> cache miss
```

Expiration is lazy during normal reads. An expired entry that nobody accesses can remain in memory, so `CleanupExpired` provides an explicit scan for removing such entries.

## 5. Concurrency model

All operations use one mutex because they can change shared state:

- `Set` changes values, links, and capacity;
- `Get` changes recency order;
- `Delete` removes nodes;
- `CleanupExpired` scans and removes nodes;
- `Len` reads the map length.

The implementation favors clear ownership and correct map/list updates over lock-free optimization.

## 6. Background cleanup lifecycle

`StartCleanup(ctx, interval)` starts one cleanup goroutine. On every tick it removes expired entries. When `ctx.Done()` is closed, the goroutine stops and its ticker is released.

```text
cache.StartCleanup(ctx, interval)
              |
              v
       cleanup goroutine
              |
       periodic map scan
              |
       ctx cancellation -> stop
```

Each cleanup scan is O(n), where n is the number of cached entries. A shorter interval reclaims memory sooner but holds the cache mutex more frequently.

## 7. Complexity

For normal operations:

```text
Get       O(1)
Set       O(1)
Delete    O(1)
LRU move  O(1)
Eviction  O(1)
```

TTL cleanup is different:

```text
lazy expiration on a known key: O(1)
full background cleanup scan:    O(n)
```

## 8. Demonstrated behavior

The runnable demonstration verifies:

- a capacity-two cache evicts the least recently used entry;
- reading an entry changes which entry is evicted;
- deleting an entry removes it from future lookups;
- TTL entries become misses after expiration;
- concurrent goroutines can use the cache safely;
- background cleanup removes expired entries that are not read.
