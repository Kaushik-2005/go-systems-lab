# Consistent Hashing

## Start with the idea

Imagine you have three cache servers and a bunch of values that need to be stored somewhere. Before reading or writing a value, you need to decide which server owns it.

That is what consistent hashing helps with.

## What is a key?

A key is just the identifier for some data. For example:

```text
user:42
session:abc123
product:987
image:profile-kaushik
```

The key is not the data itself. It is the name used to find the data.

For example:

```text
key:   user:42
value: user profile data
```

The router hashes `user:42` and uses that result to decide which cache or storage node should handle it. If the node set stays the same, the same key should keep going to the same node.

## Why not just use modulo?

The obvious approach is:

```text
node index = hash(key) % number of nodes
```

This works while the number of nodes stays fixed. The problem appears when a node is added or removed.

With three nodes, a key might map to index `1`. Add a fourth node and the divisor changes from three to four. Many keys now calculate different indexes and move to different nodes.

That is especially painful for caches because a node change can suddenly turn into a large number of cache misses.

## The ring idea

Consistent hashing puts both nodes and keys on the same circular hash space:

```text
key -> hash -> position on ring -> first node clockwise
```

```text
                         key hash
                            |
                            v
                 +-----------------------+
                 |                       |
             node-a                     node-b
                 ^                       |
                 |                       v
                 +------- node-c <-------+
```

The lookup rule is simple: hash the key, move clockwise, and choose the first node you meet.

### A small example

```text
hash space: 0 ------------------------ 100

node-a: 10        node-b: 45        node-c: 80
```

- a key at position `20` goes to `node-b`;
- a key at position `60` goes to `node-c`;
- a key at position `90` wraps around and goes to `node-a`.

The closest node is not chosen by normal numeric distance. The rule is always “first node clockwise.”

## What happens when nodes change?

When a node is added, only the key ranges immediately before its new position move to it. Most keys keep their existing owners.

When a node is removed, its ranges move to the next clockwise node. Again, keys outside those ranges stay where they were.

That is the main advantage over modulo routing: changing the cluster does not reshuffle the entire keyspace.

## Virtual nodes

One point per physical node can still produce uneven ranges. A node might get an accidentally huge section of the ring.

Virtual nodes give each physical node many positions:

```text
node-a#0  node-a#1  node-a#2  ...  node-a#99
node-b#0  node-b#1  node-b#2  ...  node-b#99
node-c#0  node-c#1  node-c#2  ...  node-c#99
```

The ring stores all those points, but each point still belongs to a physical node. More points usually spread ownership more evenly.

There is a trade-off: more virtual nodes improve distribution, but they also mean more ring metadata and more points to search.

## What this component does

This implementation answers one question:

```text
Which node should receive this key?
```

It does not answer these other questions:

```text
How many copies of the key should exist?
What happens if the selected node is down?
How should values be migrated?
Are reads strongly consistent?
```

Those are separate replication, failure-handling, migration, and consistency problems.

## What is implemented

- modulo routing as a baseline comparison;
- a sorted hash ring;
- clockwise lookup;
- adding nodes;
- removing nodes;
- configurable virtual nodes;
- demonstrations that count moved keys.

## Project structure

```text
ring/              modulo router and hash ring
cmd/demo/          runnable comparison
go.mod             Go module definition
design.md          detailed design and experiments
```

## Run it

Run from the `consistent-hashing` directory:

```powershell
go run ./cmd/demo
```

The demo does three things:

1. shows how modulo routing behaves when a node is added;
2. adds a node to a virtual-node ring and counts moved keys;
3. removes a node and counts keys that move to the next clockwise owner.

The output prints a few example key mappings and totals for a 100-key sample. The exact totals depend on the hash function, node names, and number of virtual nodes, but only a subset of keys should move when the ring changes.
