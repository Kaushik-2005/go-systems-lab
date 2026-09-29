# Key-Value Store Design

## The problem

The store maps string keys to values and needs to keep those values after a process restart.

```text
Put/Delete -> append record -> update memory
restart    -> replay log    -> rebuild map
```

The in-memory map makes reads fast. The append-only file makes mutations recoverable.

## Records, recovery, and compaction

Each JSON record contains an operation, key, and optional value. `Put` appends a put record before updating the map. `Delete` appends a tombstone and removes the key from memory.

During `Open`, records are replayed in order:

```text
PUT language=Go
PUT database=systems
DELETE language
```

The resulting map contains only `database`. Compaction writes the live map entries to a temporary file, replaces the old log, and reopens it for appending. Deleted keys and older overwritten values disappear from the compacted file.

## Concurrency and repository design

One `sync.RWMutex` protects the map and file operations. Reads use `RLock`; mutations and compaction use the write lock. Compaction holds the lock while replacing the file so no write can interleave with the rewrite.

```text
store/store.go        map, records, replay, compaction
cmd/basic/             CRUD demo
cmd/concurrent/        synchronized goroutine demo
cmd/persistent/        writing, recovery, and compaction demo
```

The store uses local files and one process. It does not provide replication, transactions, or cross-process coordination.
