# Consistent Hashing Design

## 1. Routing model

The component maps a key to one node deterministically.

```text
key
 |
 v
hash(key)
 |
 v
position on hash space
 |
 v
owner node
```

The same key and ring state produce the same node. This makes the ring useful for partitioning cache entries or routing requests to storage shards.

## 2. Why modulo routing remaps heavily

The baseline computes:

```text
hash(key) % node_count
```

With three nodes, a key might map to index `1`. After adding a fourth node, the same hash is divided by four instead of three and may map to a different index. This calculation changes for most keys, even though only one node was added.

The modulo router is included as a comparison baseline, not as the final routing mechanism.

## 3. Hash-ring representation

The ring uses a sorted slice of hash positions and maps each position to its node owner:

```text
points: [p1, p2, p3, p4, ...]
owners: point -> physical node
```

Nodes and keys are hashed into the same 32-bit space. The space is circular, so the largest position connects back to the smallest position.

## 4. Lookup algorithm

For `Lookup(key)`:

1. hash the key into a 32-bit point;
2. binary-search the sorted point slice;
3. choose the first point greater than or equal to the key point;
4. wrap to index zero if the key is beyond the final point;
5. return the owner of that point.

```text
node-a       key       node-b              node-c
  |-----------|----------|-------------------|
              |
              `-- first node clockwise: node-b
```

Lookup is O(log n) for n ring points.

## 5. Adding and removing nodes

### Add

When a node is added, its point is inserted into the sorted ring. Only keys in the range ending at the new point change owners. Other key ranges remain unchanged.

### Remove

When a node is removed, all of its points are deleted. Keys that used those points move to the next clockwise point. Other keys keep their owners.

The demo measures both operations by mapping 100 keys before and after the ring change.

## 6. Virtual nodes

With one point per physical node, the ring may contain large uneven ranges. A configurable replica count creates multiple points for each physical node:

```text
hash("node-a#0") -> point
hash("node-a#1") -> point
hash("node-a#2") -> point
```

The ring stores every virtual point but returns the physical node as the owner. More points usually improve distribution because ownership is spread across many smaller ranges.

Removing a physical node removes every virtual point associated with it.

## 7. Concurrency and state ownership

The current ring is an in-memory, single-owner data structure. It does not use a mutex yet because the demonstrations build the ring before performing lookups. If the ring becomes mutable while requests are running, node updates and lookups will need synchronization or an immutable-snapshot strategy.

The ring owns placement metadata only. It does not own the data stored by the selected node.

## 8. Demonstrated behavior

The demo uses 100 virtual points per physical node and verifies:

- modulo routing moves most keys after a node-count change;
- adding a node moves only a subset of ring-routed keys;
- removing `node-b` moves its ranges to clockwise successors;
- keys outside the changed ranges keep their original owners.

## 9. Current guarantees and limits

- lookup is deterministic for a fixed ring;
- node changes do not automatically replicate or migrate data;
- virtual nodes improve distribution but do not guarantee perfect balance;
- there is no node health checking or failure detection;
- there is no replication, quorum, consistency, or persistence mechanism;
- the ring is local to one process.
