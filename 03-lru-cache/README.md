# LRU Cache

## What is a cache?

A cache keeps frequently used data in memory so the application does not repeat slower work such as a database query, file read, network request, or expensive calculation.

```text
request -> cache.Get(key)
              |       |
             hit     miss
              |       |
           value   load and store
```

Memory is finite, so the cache needs a rule for removing entries. LRU means Least Recently Used: remove the entry that has gone unused for the longest time.

## How LRU works

The cache keeps entries in recency order:

```text
most recently used                         least recently used
          |                                           |
          v                                           v
       [cache] <-> [language] <-> [database] <-> [oldest]
```

Every successful `Get` moves an entry to the front. When capacity is full, `Set` removes the entry at the back.

## Why the implementation uses a map and a list

A map finds a key quickly but does not remember usage order. A linked list remembers order but would be slow to search by key.

```text
map["language"] --------> list node

front <-> newest <-> ... <-> oldest <-> back
```

The map gives O(1) lookup. The custom doubly linked list gives O(1) movement, deletion, and eviction.

## Features

- `Get`, `Set`, and `Delete`
- fixed capacity
- LRU eviction
- O(1) lookup and recency updates
- per-entry TTL with `SetWithTTL`
- lazy expiration during `Get`
- background expiration cleanup
- concurrency-safe access

## Project structure

```text
cache/
  cache.go            map, linked list, TTL, and cleanup
cmd/basic/
  main.go             runnable demonstration
go.mod
design.md             data structures and concurrency design
```

## Run it

Run from the `03-lru-cache` directory:

```powershell
go run ./cmd/basic
```

Example output:

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

The output shows the important behavior:

- reading `language` moves it ahead of `database`;
- inserting `cache` evicts `database`, the least recently used entry;
- `Delete` removes an entry from both the map and list;
- TTL makes an entry expire;
- concurrent callers can use the cache safely;
- background cleanup removes an expired entry that nobody reads.
