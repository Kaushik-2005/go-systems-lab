# Distributed Cache Design

## The idea

A cache avoids repeatedly reading slow storage or calling an expensive service. A distributed cache spreads entries across several cache nodes so one process does not hold every value.

The local node owns the LRU and TTL behavior. The cluster owns node membership and key placement.

## Local cache node

Each node stores a map from key to entry and a custom doubly linked list. The list head is most recently used and the tail is least recently used. A mutex protects the map and list together.

Get removes expired values lazily and moves live values to the front. Set updates recency and evicts the tail when capacity is exceeded.

## Cluster routing

Cluster contains named cache nodes and a hash ring with configurable virtual points per node. NodeFor hashes a key and chooses the first ring point at or after that hash, wrapping to the first point when necessary.

Set and Get first choose a node, then call that node's local cache operation. The cluster lock protects the node map and ring, while each node has its own lock for values. A configurable replication factor selects distinct nodes clockwise from the primary ring position.

## Node changes

Adding or removing a node rebuilds the ring. Consistent hashing limits the number of keys whose route changes compared with hash modulo routing.

Writes are sent to the primary and its replica nodes. Routing skips unhealthy nodes, so Get uses the next healthy replica when the primary is unavailable. Writes also target the currently healthy replica set. Rebalance scans live values, computes their current healthy owners, writes them to the new primary and replicas, and removes stale copies. Removing a node migrates values from that node before deleting it from the ring.

## Trade-offs

Virtual nodes improve distribution, but they add ring entries and rebuild work. A single routing process can still be a bottleneck or failure point. The cache is also intentionally in-memory, so process restart loses values.
