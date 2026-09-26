# Key-Value Store Design

## Data model

The in-memory index is:

```text
map[string]string
```

`Put` creates or replaces a value. `Get` returns a value and an existence flag. `Delete` appends a tombstone and removes the key from memory.

## Persistence flow

```text
Put/Delete
    |
    v
append JSON record to data.log
    |
    v
update in-memory map
```

The log is append-only. A delete does not erase an earlier record; it records an operation that recovery can replay.

## Recovery

`Open` scans the log from the beginning and applies each record in order:

```text
PUT language=Go
PUT database=systems
DELETE language
```

The recovered map contains only:

```text
database -> systems
```

Replay uses `applyRecord` instead of `Put` or `Delete`, so recovery updates memory without appending the same records again.

## Compaction

Append-only history grows over time. `Compact` writes the latest in-memory values to a temporary file, flushes it, replaces the old log, and reopens the file for future appends.

```text
old log:       PUT, PUT, DELETE, PUT, DELETE
compaction:    keep only live map entries
new log:       PUT database=systems
```

Keys deleted from the map are not written into the compacted file.

## Concurrency

`sync.RWMutex` protects the map and file operations. Reads use `RLock`; mutations, replay-related state changes, and compaction use the write lock. Compaction holds the lock while replacing the file so no write can interleave with the rewrite.

## Repository design

```text
store/store.go        storage state and operations
cmd/basic/             CRUD behavior
cmd/concurrent/        synchronized goroutine access
cmd/persistent/        log writing, replay, and compaction
```
