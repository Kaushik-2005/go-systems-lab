# Write-Ahead Log

## What it is

A write-ahead log (WAL) records a mutation before an application changes its state in memory. The application can use this history to recover after a crash.

## How it works

### Append and sync

![Append writes and syncs a record before it returns to the application.](diagrams/append-and-checksum.svg)

`Open(path)` creates or opens an append-only file. `Append(payload)` assigns the next sequence number. It builds a record with that sequence, the payload, and a checksum. It encodes the record as JSON with a final newline. It then writes and syncs the record.

After success, `Append` advances the next sequence and returns the record. An empty log starts at sequence zero. The payload is a string that the log does not interpret. For example, the log does not execute `set language Go`.

Apply the application mutation only after `Append` succeeds. The WAL does not change the application state itself.

### Checksum validation

The checksum is a hexadecimal SHA-224 digest of `sequence:payload`. The code calls `sha256.Sum224`. The package name includes 256, but this function calculates SHA-224.

At startup, `Open` decodes each complete record and calculates its checksum again. Invalid JSON or a different checksum causes an error. A checksum detects changes to record content if the stored checksum stays unchanged. It does not authenticate a record against a person who can change both content and checksum.

### Incomplete-tail recovery

![Startup accepts valid records, rejects corrupt complete records, and removes an incomplete final fragment.](diagrams/startup-recovery.svg)

A newline marks a complete record. If the final bytes have no newline, `Open` treats them as an incomplete tail. It removes those bytes and reopens the file for further writes. This rule also removes valid JSON if its final newline is missing.

Startup validates complete records and finds the highest sequence. It sets the next sequence to that value plus one. After sequences zero, one, and two, the next append uses three. Startup does not require historical sequences to be contiguous or strictly increasing.

`Open` validates history but does not provide a replay iterator. It does not return recovered application state. The application must separately read and apply its payload history.

## Concurrency, cost, and limits

One mutex protects appends, sequence allocation, sync, and close. Concurrent append calls run one at a time. `Close` syncs and closes the file.

Append cost includes JSON encoding, a file write, and a filesystem sync. Startup reads the entire file into memory. Thus, recovery time and temporary memory increase with file size. The component does not provide log segments, compaction, replication, or consensus.

## Project structure

```text
wal/wal.go        record format, append, validation, and tail recovery
cmd/demo/main.go  append and sequence demonstration
diagrams/         SVG illustrations
go.mod
```

## Run it

Open a terminal in the `08-write-ahead-log` directory.
Run these commands:

```powershell
go run ./cmd/demo
Get-Content ./wal.log
```

The demonstration adds three payloads and prints their sequences. A second run continues from the existing log. It does not create corrupt records or incomplete tails automatically.

Use a disposable log copy for corruption experiments.
