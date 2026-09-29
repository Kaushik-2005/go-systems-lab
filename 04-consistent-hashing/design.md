# Consistent Hashing Design

## The problem

The router needs to choose a node for every key. When nodes are added or removed, moving every key is expensive, especially for a cache.

Modulo routing changes the divisor when the node count changes:

```text
hash(key) % node_count
```

That can remap most keys. The ring keeps nodes and keys on the same circular hash space instead.

## The ring and lookup

The ring stores sorted hash points and their physical owners. Lookup hashes the key, binary-searches the points, and chooses the first point clockwise. If the key is after the final point, lookup wraps to the first point.

```text
key -> hash -> sorted points -> first clockwise owner
```

Adding a node changes only the ranges immediately before its points. Removing a node moves those ranges to the next clockwise points.

Virtual nodes create several points for each physical node. More points usually make ownership more even, while adding more ring metadata.

## State and repository design

The ring owns placement metadata, not the values stored by nodes. The demonstrations build the ring before lookup, so mutable ring operations are used by one owner at a time. Concurrent updates would need a mutex or immutable snapshots.

```text
ring/modulo.go       modulo comparison
ring/ring.go         points, owners, lookup, add, remove
cmd/demo/main.go     redistribution experiment
```

For a fixed ring, lookup is deterministic and runs in O(log n), where n is the number of ring points. Replication, data migration, health checking, and consistency are separate mechanisms.
