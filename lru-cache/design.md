# LRU Cache Design

## 1. System overview

The cache stores key-value pairs in memory and serves repeated reads without recomputing or refetching the value.

```text
Get(key)
   |
   +--> present: return cached value
   |
   `--> absent: report cache miss
```

The first implementation focuses only on correct basic operations. Capacity, recency, expiration, and concurrency will be introduced one at a time.

## 2. Core operations

- `Set(key, value)` inserts or replaces a value;
- `Get(key)` returns a value and whether it exists;
- `Delete(key)` removes a value if present.

The initial version will use a map so the first implementation exposes key-value state directly before adding the LRU data structure.

## 3. Next limitation

A map provides fast lookup but does not remember access order. Without access order, the cache cannot know which entry is least recently used when capacity is reached. The next milestone will add a recency structure.
