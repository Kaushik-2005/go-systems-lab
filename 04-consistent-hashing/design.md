# Consistent Hashing Design

## What it solves

The router must choose a node for each key and keep that choice stable while the node set changes.

```text
key -> hash -> ring position -> owner node
```

This is useful for routing cache entries, partitions, or requests to storage nodes.

## Modulo comparison

The simple baseline is:

```text
hash(key) % node_count
```

When `node_count` changes, the divisor changes for every key. Adding one node can therefore remap most keys.

The repository keeps this algorithm in `ring/modulo.go` so the redistribution difference can be measured against the ring.

## Ring representation

The ring hashes nodes and keys into the same 32-bit space. It stores sorted points and the physical owner of each point:

```text
points: [p1, p2, p3, p4, ...]
owners: point -> node name
```

The space is circular. The point after the largest position is the smallest position.

## Lookup

For `Lookup(key)`:

1. hash the key;
2. binary-search the sorted point slice;
3. choose the first point greater than or equal to the key hash;
4. wrap to index zero when the key is after the final point;
5. return that point's physical owner.

```text
node-a       key       node-b              node-c
  |-----------|----------|-------------------|
              |
              +--> first node clockwise: node-b
```

With n ring points, lookup is O(log n).

## Adding and removing nodes

Adding a node inserts its points into the sorted ring. Only the ranges immediately before those points change owners.

Removing a node deletes all of its points. Keys that used those points move to the next clockwise points. Other ranges keep their owners.

The demo maps 100 keys before and after each change and counts how many owners changed.

## Virtual nodes

One point per physical node can create large, uneven ranges. A replica count creates several points per physical node:

```text
hash("node-a#0") -> point
hash("node-a#1") -> point
hash("node-a#2") -> point
```

The lookup returns `node-a`, not the virtual-node label. More virtual nodes usually improve distribution, while increasing ring metadata.

## Repository design

```text
ring/modulo.go      modulo routing baseline
ring/ring.go        sorted points, owners, lookup, add, remove
cmd/demo/main.go     redistribution experiment
```

The ring owns placement metadata. It does not store values, replicate data, check node health, migrate data, or provide consistency between nodes.

## Concurrency

The demonstrations build the ring before performing lookups, so the ring is used by one owner at a time. If node changes and lookups happen concurrently, the ring needs a mutex or an immutable snapshot strategy around its sorted points and owners.

## Guarantees

For a fixed ring, the same key maps deterministically to the same node. Adding or removing a node changes only the affected clockwise ranges. Virtual nodes improve distribution but do not guarantee perfect balance.
