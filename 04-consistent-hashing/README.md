# Consistent Hashing

## What is it?

Consistent hashing decides which node should own a key.

A key is the identifier for some data:

```text
user:42
session:abc123
product:987
image:profile-kaushik
```

The key is not the data itself. It is the name used to find the data. The router hashes the key and maps the result to a cache server, storage node, or shard.

## Why not `hash(key) % N`?

Modulo routing uses the number of nodes as the divisor:

```text
node = hash(key) % number_of_nodes
```

With three nodes, a key may map to index 1. Add a fourth node and the divisor changes, so many keys calculate new indexes. A cache then experiences many misses even though only one node changed.

## The ring

Consistent hashing puts nodes and keys on one circular hash space:

```text
hash(key) -> position -> first node clockwise
```

```text
                 node-a
                /      \\
             key        node-b
                \\      /
                 node-c
```

If a key lands after node-a and before node-b, node-b owns it. If it lands after the last node, the search wraps around to the first node.

## Virtual nodes

One point per physical node can create uneven ranges. Virtual nodes give each physical node many positions:

```text
node-a#0  node-a#1  node-a#2  ...
node-b#0  node-b#1  node-b#2  ...
node-c#0  node-c#1  node-c#2  ...
```

The ring stores every virtual point but returns the physical node as the owner. More points usually spread keys more evenly.

## What node changes do

Adding a node moves the key ranges immediately before its new points. Removing a node moves those ranges to the next clockwise owners. Keys outside those ranges keep their owners.

That is the main advantage over modulo routing: changing the node set moves a subset of keys instead of reshuffling most of the keyspace.

## Project structure

```text
ring/
  modulo.go          modulo baseline
  ring.go            hash ring, lookup, node changes, virtual nodes
cmd/demo/
  main.go            routing and redistribution demonstration
go.mod
design.md             ring data structure and lookup reasoning
```

## Run it

Run from the `04-consistent-hashing` directory:

```powershell
go run ./cmd/demo
```

The demo maps 100 keys, adds a node, counts moved keys, removes a node, and counts moved keys again. It also prints example mappings before and after each change.

The ring answers this question:

```text
Which node should receive this key?
```

Replication, data migration, node health, and consistency are separate mechanisms.
