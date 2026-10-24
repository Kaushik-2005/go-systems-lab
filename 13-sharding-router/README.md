# Sharding Router

## What it is

A sharding router assigns keys to storage partitions called shards. The router stores placement metadata. A separate shard map stores values in a map for each shard.

## How it works

### Hash-based routing

![The key hash selects a shard. Put and Get use the selected shard's map.](diagrams/hash-routing.svg)

`Router` copies an ordered shard list. `Route(key)` calculates `FNV-1a(key) % shardCount` and selects the shard at that index. An empty list returns `ErrNoShards`. A key keeps the same assignment while shard membership and order stay unchanged.

`ShardMap.Put` selects the shard before it writes the value. `Get` selects the shard again and reads that shard's map. It reports the result and the selected shard.

Modulo selection takes O(1) time after the key hash calculation. Hash routing distributes keys across a fixed list. It does not group keys by their string values.

### Range-based routing

![Half-open ranges include their start boundary and exclude their end boundary.](diagrams/range-routing.svg)

`RangeRouter` uses half-open string intervals. Start is inclusive. End is exclusive. An empty End represents infinity.

For `[a, m)`, `[m, t)`, and `[t, infinity)`, apple selects the first range. Profile selects the second range. User selects the third range. A key equal to `m` selects the second range.

Construction copies and sorts the intervals. `Route` checks them in order and selects the first interval that contains the key. An uncovered key returns `ErrInvalidRanges`. Use intervals without overlap. Use an empty End only for the final interval.

The demonstration starts the first interval at an empty string. This includes keys before `a`. Range lookup takes O(r) time for r intervals. Nearby keys stay together, but one popular interval can receive excessive load.

The current `ShardMap` uses the hash router. The range router has a separate demonstration. It does not connect to this shard map.

### Membership changes and migration

![Shard addition changes the route before Rebalance moves the stored value.](diagrams/shard-rebalancing.svg)

`AddShard` adds a shard at the end of the list. A duplicate returns `ErrShardExists`. `RemoveShard` removes a known shard. An unknown shard returns `ErrShardNotFound`. A changed list changes the modulo divisor and can move many key assignments.

Membership changes do not move values automatically. `ShardMap.Rebalance` creates new buckets for the current shards. It routes each stored key again and copies its value to the new bucket. It then replaces the bucket map. It returns the number of records with different owners.

Before migration, a changed route can cause a read miss. The value can still exist in an old bucket. Rebalance takes O(k + s) time for k records and s shards, plus hash and allocation costs. It needs temporary bucket space. It blocks shard-map reads and writes during the copy.

## Concurrency and limits

The router uses `sync.RWMutex` for the shard list. The shard map has a separate mutex for values. A route and the subsequent value access do not share one membership snapshot. Coordinate membership changes and migration externally during concurrent traffic.

All values exist in one process. The component has no networking, replication, persistent storage, or online migration protocol. Component four provides consistent hashing, which reduces changed assignments during membership changes.

## Project structure

```text
shard/router.go   deterministic hash routing and membership
shard/range.go    interval-based routing
shard/map.go      per-shard maps and rebalancing
cmd/demo/main.go  routing and migration demonstration
diagrams/         SVG illustrations
go.mod
```

## Run it

Open a terminal in the `13-sharding-router` directory.
Run this command:

```powershell
go run ./cmd/demo
```

The demonstration stores eight keys and prints their shard assignments. It adds shard-d, rebalances records, and reads user-1 again. It then removes shard-d and repeats rebalancing. Output separates `routes_moved` from `records_rebalanced`. A separate example routes apple, profile, user, and zebra through string intervals.
