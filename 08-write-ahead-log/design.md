# Write-Ahead Log Design

## What it solves

A state change can be lost if a process crashes before the change is saved. A write-ahead log records the change first:

```text
mutation -> append and sync WAL -> apply to memory
```

On restart, replay reconstructs the state in the same order.

## Record model

Each JSON record contains:

```text
sequence + payload + checksum
```

The sequence number gives a total order. The checksum is calculated from the sequence and payload, so recovery can reject a complete record whose contents changed.

## Recovery

`Open` reads complete newline-delimited records, validates their checksums, and sets the next sequence number after the highest valid record.

An incomplete final line is treated as a partial write. The valid prefix is kept, the incomplete tail is truncated, and the file is reopened for appending. A complete record with an invalid checksum returns an error instead of being replayed.

## Concurrency and durability

`Log.mu` serializes appends, sequence allocation, sync, and close. `Append` writes the record and calls `file.Sync` before returning, making the durability step explicit.

## Repository design

```text
wal/wal.go        log state and record operations
cmd/demo/main.go  append and recovery demonstration
```
