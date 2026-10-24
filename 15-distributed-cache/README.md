# Distributed Cache

## What it is

This component combines local caches, consistent hashing, replication, and migration. Each named node is an object in one process. The nodes are not separate network servers.

## How it works

### Local cache nodes

![Reads change local recency order. Insertion can remove the least recently used entry.](diagrams/local-node-cache.svg)

Each `Node` has a key map and a custom doubly linked list. The head contains the most recently used entry. The tail contains the least recently used entry. This order provides least recently used (LRU) eviction.

`Set` inserts or changes an entry at the head. It removes the tail if entry count exceeds capacity. `Get` removes an expired entry or moves a valid entry to the head. `Delete` removes an entry from the map and list.

A positive time to live (TTL) sets an expiration time. Zero or a negative TTL means no expiration. `Entries` removes expired entries and returns a key/value snapshot. `Len` can include expired entries until a read or scan removes them. The node has no background expiration loop.

### Placement, replicas, and health

![Set writes selected replicas. Get reads the first healthy owner in ring order.](diagrams/replication-and-health.svg)

`NewCluster(names, capacity, replicas, replicationFactor)` creates named nodes and a sorted FNV-1a hash ring. The `replicas` parameter specifies virtual points per node. The `replicationFactor` parameter specifies physical copies per value. Both counts and capacity must be positive.

Routing hashes the key and finds the first ring point at or after that hash. It then moves clockwise through the ring. It collects distinct healthy owners up to the requested replication factor. It returns to the first point when necessary. Multiple points for one node do not create additional copies.

If too few nodes are healthy, routing returns the available subset. If none are healthy, it returns `ErrNoNodes`. `Set` writes to selected owners one at a time and returns the first owner's name. `Get` reads only the first selected healthy owner. It does not try other replicas after a miss.

`SetNodeHealth` changes a manual flag. It does not run a health probe. If the primary is unhealthy, routing selects another healthy owner. A previous replica can then serve the read. New writes use the current healthy set.

A recovered node can contain an old value. The system does not compare versions or repair copies during reads. Thus, health recovery does not establish that a replica has the newest value.

### Membership changes and rebalancing

![Rebalance copies values to current owners and removes copies from other nodes.](diagrams/cache-rebalancing.svg)

`AddNode` creates an empty node and rebuilds the ring. It does not move existing values automatically. A changed route can cause a miss before rebalancing.

`Rebalance` scans nodes in name order and collects live values. It calculates current healthy owners and writes the selected copies. It then removes copies from other nodes. If copies disagree, the first scanned value wins. The code has no versions to select the newest copy.

The moved count compares the first scanned source with the new primary. It does not count every replica write. `RemoveNode` first copies that node's live entries. It then removes membership and rebuilds the ring. It writes those entries to remaining selected owners under the membership write lock.

The cluster rejects removal of its last node. Removal operates on local objects. It does not model data recovery from an unavailable network server.

Migration writes use TTL zero. They do not preserve the original expiration time. Moved values therefore have no expiration. Capacity limits can also cause eviction during migration.

## Concurrency, cost, and limits

The cluster's `sync.RWMutex` protects membership, health flags, and ring metadata. Each node has a mutex for values and recency order. `Set` and `Get` release the membership read lock after node selection. They access values afterward. Their membership selection and value access do not form one atomic operation.

Rebalance holds a membership read lock. Cache writes can still run concurrently. Its scans and copies do not form a transactional snapshot. Coordinate migration with writes when the newest value must survive.

Local lookup and eviction take O(1) time on average. Routing starts with O(log p) binary search for p ring points. It can then check the whole ring for distinct healthy owners. Rebalance cost increases with entry count and replication factor.

A restart loses all values. The component has no network transport, durable storage, consensus, quorum protocol, automatic health checks, or conflict resolution.

## Project structure

```text
cache/node.go     local LRU list, TTL, and snapshots
cache/cluster.go  hash ring, replicas, health, and migration
cmd/demo/main.go  local-node and multi-node demonstrations
diagrams/         SVG illustrations
go.mod
```

## Run it

Open a terminal in the `15-distributed-cache` directory.
Run this command:

```powershell
go run ./cmd/demo
```

The demonstration shows local eviction and TTL expiration. It creates three nodes with 20 virtual points per node and replication factor two. It writes user-1 and marks its primary unhealthy. It then adds cache-d and rebalances values. Finally, it removes cache-d and reads user-1 again. Output shows selected owners, changed routes, and moved records.
