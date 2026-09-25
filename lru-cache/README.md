# LRU Cache

## What is a cache?

A cache keeps recently used data in memory so repeated reads can avoid slower work such as a database query, file read, network request, or expensive computation.

```text
request
   |
   v
cache.Get(key)
   |
   +--> hit: return the value immediately
   |
   `--> miss: load the value, store it, return it
```

Memory is limited, so a cache needs an eviction policy. LRU means Least Recently Used: when the cache is full, remove the entry that has gone unused for the longest time.

## How LRU works

The cache keeps entries in recency order:

```text
most recently used                              least recently used
        |                                                  |
        v                                                  v
     [language] <-> [cache] <-> [database] <-> ... <-> [oldest]
```

Every successful `Get` moves the entry to the front. When `Set` exceeds capacity, the entry at the back is evicted.

## Why two data structures are needed

A map gives fast lookup but does not remember order. A linked list remembers order but is slow to search by key.

```text
map:             "language" ───────> list entry
                                      |
list:       [language] <-> [cache] <-> [database]
                front                         back
                newest                        oldest
```

The map provides O(1) lookup. The custom doubly linked list provides O(1) movement to the front and removal from the back.

## Features implemented

- `Get`, `Set`, and `Delete`;
- fixed capacity;
- LRU eviction;
- O(1) lookup, recency updates, and eviction;
- optional per-entry TTL through `SetWithTTL`;
- lazy expiration during `Get`;
- concurrency-safe access;
- optional background expiration cleanup with context cancellation.

## Project structure

```text
cache/            cache state, linked list, TTL, and cleanup
cmd/basic/        runnable demonstration
go.mod            Go module definition
design.md         detailed design and behavior
```

## Run the demonstration

Run from the `lru-cache` directory:

```powershell
go run ./cmd/basic
```

The demonstration prints the recency order after each operation and shows:

1. inserting two entries into a capacity-two cache;
2. reading one entry to make it recent;
3. inserting a third entry and evicting the untouched entry;
4. deleting an entry;
5. expiring an entry through TTL;
6. concurrent access from multiple goroutines;
7. background removal of expired entries.

### Example output explained

```text
after Set(language)      order(MRU->LRU)=[language] size=1
after Set(database)      order(MRU->LRU)=[database language] size=2
after Get(language)      order(MRU->LRU)=[language database] size=2
after Set(cache)         order(MRU->LRU)=[cache language] size=2
eviction check: language="Go" found=true database_found=false
after Delete(language)   order(MRU->LRU)=[cache] size=1
delete check: deleted=true
ttl: found_before_expiry=true found_after_expiry=false
concurrent access: operations=1000 misses=0 entries=10
background cleanup: entries=0
```

Each line corresponds to a specific operation in `cmd/basic/main.go`:

- `after Set(language)`: `Set` inserts the first entry. It is both the most recently used and least recently used entry.
- `after Set(database)`: the new entry is placed at the front, so `database` is now MRU and `language` is LRU.
- `after Get(language)`: the successful `Get` moves `language` to the front. This is why a read changes the order.
- `after Set(cache)`: capacity is two, so inserting `cache` evicts `database`, which is at the back of the list.
- `eviction check`: `language` remains because it was recently accessed; `database` is absent because it was evicted.
- `after Delete(language)`: `Delete` removes `language` from both the map and linked list.
- `delete check`: the boolean returned by `Delete` confirms that an entry was removed.
- `ttl`: `SetWithTTL` makes `session` valid initially; after the sleep, `Get` lazily removes it and reports a miss.
- `concurrent access`: ten goroutines perform 100 operations each. Zero misses means every immediate `Set`/`Get` pair succeeded under concurrent access.
- `background cleanup`: a cleanup goroutine removes an expired entry without requiring a caller to `Get` it first.

The `order(MRU->LRU)` text comes from the cache's `Keys` method, which walks the linked list from `head` to `tail`.
