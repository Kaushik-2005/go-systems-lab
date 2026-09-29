# Write-Ahead Log

## What it is

Lets say your application changes some state and then crashes before the change is saved. A write-ahead log records the change first so it can be replayed later.

```text
mutation -> write to WAL -> apply to memory
restart  -> replay WAL    -> rebuild state
```

## How it works

The log appends ordered JSON records. Every record has a sequence number and checksum.

```text
sequence=0 -> mutation A
sequence=1 -> mutation B
sequence=2 -> mutation C
```

It also handles:

- append-only writes;
- monotonically increasing sequence numbers;
- startup replay;
- SHA-256 checksum validation;
- complete-record corruption detection;
- incomplete final-record recovery.

## Project structure

```text
wal/          record format, append, replay, and validation
cmd/demo/     append and recovery demonstration
go.mod
```

## Run it

Run from the `08-write-ahead-log` directory:

```powershell
go run .\cmd\demo\main.go
```

Inspect the log with:

```powershell
Get-Content .\wal.log
```

The demo prints sequence numbers and writes the records to `wal.log`. Corrupting a complete record triggers checksum rejection, while an incomplete final record is removed during recovery.
