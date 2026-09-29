# Key-Value Store

## What it is

Lets say your application needs to store and retrieve data using a name. A key-value store maps a key to a value.

```text
key:   language
value: Go
```

## How it works

The store keeps an in-memory map and writes every mutation to an append-only log.

```text
Put/Delete -> append record -> update memory
restart    -> replay log    -> rebuild map
```

It also handles:

- `Put`, `Get`, and `Delete`;
- JSON mutation records;
- tombstones for deletes;
- startup recovery;
- log compaction;
- concurrent access.

## Project structure

```text
store/             map, log records, replay, and compaction
cmd/basic/          CRUD demonstration
cmd/concurrent/     concurrent access demonstration
cmd/persistent/     persistence and recovery demonstration
go.mod
```

## Run it

Run from the `07-key-value-store` directory:

```powershell
go run .\cmd\basic\main.go
go run .\cmd\concurrent\main.go
go run .\cmd\persistent\main.go
```

The persistent demo writes records, reopens the log, rebuilds the map, and compacts obsolete history.
