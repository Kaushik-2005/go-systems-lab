# Write-Ahead Log

## What it is

An append-only log that records ordered mutations before a system applies them to in-memory state.

## What it solves

After a process restarts, the log can be replayed to rebuild state. Sequence numbers preserve mutation order, checksums detect corrupted records, and incomplete final records can be discarded safely.

## Project structure

```text
wal/wal.go        record format, append, replay, checksum, tail recovery
cmd/demo/main.go  runnable demonstration
go.mod
```

## Run it

```powershell
go run .\cmd\demo\main.go
```

The demo appends records and prints their sequence numbers. `wal.log` can be inspected to see the JSON records and checksums.
