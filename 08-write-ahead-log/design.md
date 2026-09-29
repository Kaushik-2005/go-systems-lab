# Write-Ahead Log Design

## The problem

A process can crash after a mutation is created but before in-memory state is safely recoverable. The WAL records the mutation first:

```text
mutation -> append and sync WAL -> apply to memory
restart  -> replay WAL          -> rebuild state
```

## Records and recovery

Each JSON record contains a sequence number, payload, and checksum. The sequence number preserves order. The checksum is calculated from the sequence and payload, so recovery can reject a complete record that was corrupted.

`Open` reads complete newline-delimited records, validates them, and starts the next sequence after the highest valid record. If the final line is incomplete, the valid prefix is kept, the partial tail is truncated, and the file is reopened for appending.

```text
valid record       -> replay
complete bad record -> checksum error
incomplete tail    -> truncate and continue
```

## Concurrency and repository design

One mutex protects appends, sequence allocation, sync, and close. `Append` writes the record and calls `file.Sync` before returning.

```text
wal/wal.go        log state, record format, replay, and validation
cmd/demo/main.go  append, corruption, and tail-recovery experiments
```

The WAL is a local file mechanism. It does not apply mutations to an application state machine by itself, replicate records, or implement consensus.
