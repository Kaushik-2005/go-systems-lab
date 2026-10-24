# Key-Value Store

## What it is

A key-value store associates a string key with a string value. This implementation keeps values in memory. An optional append-only log preserves mutations for recovery.

## How it works

`New()` creates a memory-only store. `Open(path)` adds a JSON log and reads previous records to recover values. Both modes read values from the map.

### Writes and tombstones

![A persistent mutation writes and syncs its record before it changes the map.](diagrams/persistent-write-path.svg)

`Put(key, value)` encodes a JSON record and adds a newline. It writes the record and calls `file.Sync`. After success, it changes the map. If the write or sync fails, it returns an error before the map change.

`Delete(key)` writes a delete record before it removes an existing key. This record is a tombstone. A missing key returns `false` without a record. `Get(key)` reads only the map and reports whether the key exists.

```json
{"operation":"put","key":"language","value":"Go"}
{"operation":"delete","key":"language"}
```

The log records each mutation without a full database rewrite. Old values and tombstones remain until compaction. Map access takes O(1) time on average. Persistent mutations also require encoding, file writes, and sync operations.

### Recovery by replay

![Replay applies puts and deletes in order to rebuild the live map.](diagrams/log-recovery.svg)

`Open(path)` creates or opens the log file. It reads newline-delimited JSON records and applies them in order to an empty map. A later put replaces an earlier value. A tombstone removes its key during replay.

For `PUT language=Go`, `PUT database=systems`, then `DELETE language`, recovery leaves only `database=systems`. The file stays open for future writes. Replay time increases with the full log size, including obsolete records.

Invalid JSON or a scanner error causes startup to fail. This component does not repair a partial final record or check record checksums. The write-ahead-log component provides those mechanisms.

### Log compaction

![Compaction keeps one put record for each live key.](diagrams/log-compaction.svg)

`Compact()` holds the write lock while it replaces the log. It sorts live keys and writes one put record per key to a temporary file. The temporary file is in the same directory as the log.

The function flushes the buffer and syncs the temporary file. It closes that file and the old log. It renames the temporary file over the old log. It then reopens the log for further writes.

For example, the log can contain two language values and a deleted temporary key. Compaction keeps only the latest language value. Replay then gives the same live map. Compaction of a memory-only store has no effect.

The sort takes O(k log k) time for k live keys. Compaction also requires time to write their values. It reduces file size and future replay work. It blocks reads and mutations during the rewrite. It does not provide a transactional crash-recovery protocol.

## Concurrency and limits

One `sync.RWMutex` protects map and file operations. Reads use `RLock`. Mutations, compaction, and close use the write lock. The lock prevents a concurrent mutation during log replacement. `Close()` syncs and closes an open file.

All live values must fit in memory. The store uses one process and local files. It does not provide replication, transactions, or coordination between processes.

## Project structure

```text
store/             map, log records, replay, and compaction
cmd/basic/          CRUD demonstration
cmd/concurrent/     concurrent access demonstration
cmd/persistent/     persistence and recovery demonstration
diagrams/           SVG illustrations
go.mod
```

## Run it

Open a terminal in the `07-key-value-store` directory.
Run these commands:

```powershell
go run .\cmd\basic\main.go
go run .\cmd\concurrent\main.go
go run .\cmd\persistent\main.go
```

The persistent demonstration writes records, compacts the log, and closes the store. It then reopens the log to recover the live map.
