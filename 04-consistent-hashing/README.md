# Consistent Hashing

## What it is

Lets say you have several cache or storage servers and need to decide where each key should go. Consistent hashing maps a key to a node while reducing movement when nodes are added or removed.

```text
key -> hash -> ring position -> node
```

## How it works

Nodes and keys are placed on the same circular hash ring. The lookup hashes the key, moves clockwise, and chooses the first node it reaches.

```text
node-a -> key -> node-b -> node-c
  ^                         |
  +-------------------------+
```

Virtual nodes give each physical node multiple positions, which makes key distribution more even.

The component also includes a modulo baseline so the difference in key movement can be observed.

## Project structure

```text
ring/       modulo router and consistent-hash ring
cmd/demo/   key redistribution demonstration
go.mod
```

## Run it

Run from the `04-consistent-hashing` directory:

```powershell
go run ./cmd/demo
```

The demo maps keys, adds a node, counts moved keys, removes a node, and counts moved keys again.
