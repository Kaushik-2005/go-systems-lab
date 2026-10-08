# Distributed Cache

## What it is

A distributed cache keeps frequently used data close to the services that need it. This component now has multiple in-memory cache nodes behind a consistent-hash router.

text
application -> consistent-hash router -> cache-a
                                      -> cache-b
                                      -> cache-c

Each node has its own custom LRU cache with TTL expiration. The router decides which node owns a key. With replication factor 2, the key is written to the primary and the next distinct node on the ring.

## How it works


- A key is hashed onto a ring.
- Virtual points on the ring improve distribution.
- The router walks clockwise to find the node for the key.
- Set and Get use the same routing decision.
- Each selected node applies its own LRU and TTL rules.
- Adding a node changes only a subset of key placements.

The demo prints the replica nodes for user-1:

user-1 replicas: [cache-b cache-a]

The first node is the primary used for reads. Both nodes receive writes. When cache-b is marked unhealthy, the router skips it and the read comes from cache-a:

The demo also prints the routes before and after adding cache-d:

after adding cache-d: moved=4/6 records_migrated=1
after removing cache-d: value="Kaushik" found=true node=cache-b

after primary failure: value="Kaushik" found=true node=cache-a

The exact number depends on the keys and virtual-node layout. Existing values are not migrated yet, so a key that moves can temporarily miss. Replication and migration are both present. Rebalancing copies live values to their new primary/replica owners and removes stale copies. Removing a node migrates its values before the node leaves the cluster.

## Run it

go run ./cmd/demo

Example output:

get language: value="Go" found=true
after capacity eviction: database_found=false
ttl: before_expiry=true after_expiry=false
cluster nodes: [cache-a cache-b cache-c]
cluster get: value="Kaushik" found=true node=cache-b
after adding cache-d: moved=4/6 records_migrated=1
after removing cache-d: value="Kaushik" found=true node=cache-b

## Project structure

text
15-distributed-cache/
├── cache/node.go         # one concurrency-safe LRU + TTL cache node
├── cache/cluster.go      # virtual-node ring and cluster routing
└── cmd/demo/main.go      # local node and multi-node demonstration
