# Sharding Router

## What it is

A sharding router decides which storage node should handle a key. Instead of keeping every key on every node, the data is split across shards.

```text
client
  |
  v
router -- user-1  --> shard-a
       -- user-2  --> shard-c
       -- order-1 --> shard-b
```

The first version uses a hash of the key and maps it to one of the configured shards.

## How it works

1. The router receives a key.
2. It hashes the key with FNV-1a.
3. It takes the hash modulo the number of shards.
4. The resulting index selects the shard.

The same key gets the same shard as long as the shard list stays unchanged. The demo prints the mapping and repeats `user-1` to make that property visible.

## Shard map

The shard map adds storage to the router. It keeps a separate in-memory key/value map for each shard. A Put first asks the router for the owner shard, then stores the key in that shard. A Get repeats the same routing decision and reads from that shard.

The demo begins with:

stored user-1: value="Kaushik" found=true shard=shard-b

This shows both parts working together: the value is stored and the router reports where it lives.

## Range-based routing

The demo also creates ordered ranges:

[a, m) -> range-a
[m, t) -> range-b
[t, infinity) -> range-c

A key is routed by comparing its value with the range boundaries. For example, apple goes to range-a, profile goes to range-b, and user goes to range-c.

Hash routing spreads keys without understanding their values. Range routing keeps nearby keys together, which is useful for ordered scans and range queries. Its trade-off is that one range can become much hotter than the others.

## Add and remove shards

The demo starts with three shards, records where eight keys go, adds shard-d, and routes the same keys again. It prints how many keys moved:

after adding shard-d: moved=3/8

The exact keys depend on the hash and shard names, but the important point is that changing the shard count changes the modulo divisor. A key that was on one shard may now point to another shard, so its data would need to move too.

The demo then removes shard-d and prints another redistribution count. This is why simple modulo hashing is easy to understand but awkward when nodes change.

## Project structure

```text
13-sharding-router/
├── shard/router.go       # deterministic key-to-shard routing
├── shard/map.go          # per-shard in-memory key/value maps
└── cmd/demo/main.go      # routing demonstration
```

## Run it

```bash
go run ./cmd/demo
```

Example output:

```text
before adding shard-d:
key=cart-1 shard=shard-b
key=cart-2 shard=shard-b
key=order-1 shard=shard-b
key=order-2 shard=shard-b
key=profile-1 shard=shard-a
key=user-1 shard=shard-b
key=user-2 shard=shard-b
key=user-3 shard=shard-c
after adding shard-d: moved=3/8
```

The exact shard names for most keys depend on the hash, but repeated `user-1` stays on the same shard.
