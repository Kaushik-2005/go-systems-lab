# Consistent Hashing

## What it is

Consistent hashing assigns keys to nodes. It reduces the number of keys that change node when a node enters or leaves the group.

## How it works

This component provides a modulo router and a hash ring. Both select a physical node for a key. Neither stores values. The diagrams use example positions and hash results.

### Modulo baseline

`NewModulo(nodes)` copies an ordered list of nodes. `Lookup(key)` calculates an FNV-1a hash. It selects `nodes[hash(key) % len(nodes)]`. If the list is empty, it returns an empty string.

![A change from three nodes to four changes the modulo results.](diagrams/modulo-routing.svg)

For hash result seven, three nodes select index `7 % 3 = 1`, or B. With four nodes, the same hash selects index `7 % 4 = 3`, or D. One additional node can change the assignment of many keys.

Selection takes O(1) time after the hash calculation. The router stores a fixed list. To compare two node groups, create a router for each group.

### Ring lookup

The ring stores sorted hash points and their physical owners. It also records the points for each node. The ring uses the first 32 bits of SHA-1 to hash keys and virtual-node names.

![Hash 30 selects B at 45. Hash 95 returns to A at 10.](diagrams/ring-lookup.svg)

`Lookup(key)` finds the first ring point greater than or equal to the key hash. It uses binary search. If the key hash exceeds the final point, the search selects the first point. An empty ring returns an empty string.

For points `A:10`, `B:45`, and `C:80`, hash 30 selects B. Hash 95 selects A. For a fixed ring, the same key always selects the same node. Lookup takes O(log p) time for p ring points.

The modulo router uses a different hash function. Its numeric hash results do not directly correspond to the ring's hash results.

### Adding and removing nodes

![Node D takes only the interval before its point. Virtual points divide each node's assigned intervals.](diagrams/membership-and-virtual-nodes.svg)

If D enters at position 30, hashes in `(10, 30]` change from B to D. If D leaves, those hashes return to B. The same rule applies to the interval before each virtual point.

`AddNode` ignores an existing node. For a new node, it creates points and sorts the ring. If a point already exists, the function increases that point until it finds an unused position. `RemoveNode` removes all points for the specified node. It ignores an unknown node.

These functions change assignments only. An application must move stored values separately when their assigned nodes change.

### Virtual nodes

`New(100)` creates 100 virtual points for each physical node. Names such as `node-a#0` and `node-a#1` select different points for the same physical owner. `New()` uses one point per physical node.

More points usually give a more equal distribution of intervals. They also increase metadata and sort costs. Virtual nodes do not copy values to other physical nodes. A frequently used key can still give one node much more work.

```go
placement := ring.New(100)
placement.AddNode("node-a")
placement.AddNode("node-b")
owner := placement.Lookup("user:42")
```

## State, concurrency, and limits

Ring metadata uses O(p) space. The demonstration programs prepare rings before they read keys. The ring has no mutex. Concurrent membership changes and lookups require external synchronization or immutable snapshots.

The ring does not provide replication, data migration, or health checks. A node assignment does not show that the node is available or contains the value.

## Project structure

```text
ring/       modulo router and consistent-hash ring
cmd/demo/   key redistribution demonstration
diagrams/   SVG illustrations
go.mod
```

## Run it

Open a terminal in the `04-consistent-hashing` directory.
Run this command:

```powershell
go run ./cmd/demo
```

The demonstration uses 100 virtual points per node. It compares routes for 100 keys before and after node addition or removal. It prints example routes and counts changed assignments. It does not move stored values.
