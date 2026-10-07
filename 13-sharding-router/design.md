# Sharding Router Design

## The idea

When one storage node cannot hold or serve all the data, keys can be divided among several nodes. A router hides that partitioning from the client and sends each key to its shard.

The first step is deterministic hash routing. The router stores an ordered list of shard IDs and uses the key's hash to select an index.

```text
index = hash(key) % number_of_shards
```

## State and behavior

`Router` owns a copy of the shard list. `Route` hashes a key with FNV-1a and returns the shard at the calculated index. It returns an error if the router has no shards instead of silently dropping the request.

The router does not own the values. ShardMap adds a separate in-memory map for each configured shard. Put routes first and stores in that shard; Get routes again and reads from the same shard. This keeps placement and storage as separate responsibilities.

## Range-based routing

RangeRouter stores ordered half-open intervals. A range includes its Start and excludes its End. An empty End represents infinity:

[a, m) -> shard-a
[m, t) -> shard-b
[t, infinity) -> shard-c

Route scans the ranges and returns the first interval containing the key. This keeps neighboring keys together, unlike hashing, and makes ordered key scans natural. The trade-off is that a popular range can become a hot shard.

## Adding and removing shards

AddShard and RemoveShard update the shard list under a mutex. Routing also uses a read lock, so a route never observes a partially updated list.

The demo routes the same keys before and after a shard change and counts movement. With modulo routing, changing from three shards to four changes the divisor for every key:

hash(key) % 3  ->  hash(key) % 4

ShardMap.Rebalance rebuilds the per-shard buckets using the current router. It routes every existing key again and moves each record into its new bucket. The demo reports how many records moved and reads user-1 after the move to prove the value survived.

## Trade-off

Modulo hashing is simple and spreads keys reasonably when the shard set is fixed. However, adding or removing one shard changes the divisor, so many keys move to different shards. That can cause cache misses and data movement.

This limitation motivates the next step: compare hash partitioning with consistent hashing and then demonstrate node addition, removal, and rebalancing.
