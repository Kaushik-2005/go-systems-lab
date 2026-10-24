# Replicated Key-Value Store

## What it is

This component copies a leader's key-value state to follower processes. The leader accepts writes. Followers request ordered mutation records and serve reads from their local maps.

## How it works

### Asynchronous leader / follower replication

![The leader responds before the follower requests and applies the next record.](diagrams/asynchronous-replication.svg)

Each leader `Put` or `Delete` assigns sequence `lastSeq + 1`. It changes the map and adds a record to the log in memory. Each record has a sequence, key, value, and `deleted` flag. The sequence identifies a position in the whole log. It is not a version for one key.

The leader responds after its own change. It does not wait for followers. A follower can return an older value or report that a key is missing. This delay is replication lag. If the leader fails before replication, followers can lack accepted writes.

Each follower requests `GET /records?from=N`, where N is its last applied sequence. It applies returned records in order. The default request interval is 500 ms. The HTTP client timeout is two seconds. A failed request causes a log message and a later retry.

### Sequence validation

![A follower ignores an old sequence, accepts the next sequence, and rejects a gap.](diagrams/sequence-validation.svg)

`Apply(record)` ignores a sequence at or below `lastSeq`. It accepts exactly `lastSeq + 1`. It returns `ErrOutOfOrder` for a gap. A put replaces a value. A delete removes a key.

Repeated records do not change the intended single-leader history. The code does not compare the content of records with duplicate sequences. Conflicting content at the same sequence is not an error that this check detects.

The leader supplies its retained records through `RecordsAfter`. Followers change their maps and sequences without a second replication log. They do not supply records to other followers.

### Reconnect and restart catch-up

![A disconnected follower keeps its sequence. A restarted follower rebuilds state from sequence zero.](diagrams/reconnect-and-restart.svg)

A follower process keeps its state during a network failure. If it stops at sequence two and the leader reaches four, it requests records after two. It then applies records three and four.

A restarted follower loses its map and sequence. It starts at zero and reads the leader's retained history again. A leader restart loses that history too. The system has no epoch or snapshot protocol to connect a restarted leader's new history with existing followers.

## Concurrency and limits

Each store uses a `sync.RWMutex`. Leader writes and follower apply operations take the write lock. Map reads, sequence reads, and record copies take the read lock. Follower HTTP reads can run during its replication loop.

Map access takes O(1) time on average. Catch-up copies and transmits the requested log suffix. The log has no size limit. There is no durable storage, quorum acknowledgment, automatic promotion, or leader election.

If the leader stops, followers retain their last known state. They can continue to serve reads. New writes require the leader.

## Project structure

```text
store/store.go        state, ordered records, and apply logic
cmd/leader/main.go    write API and source replication log
cmd/follower/main.go  polling replica and local read API
diagrams/             SVG illustrations
go.mod
```

## Run it

Open a terminal in the `12-replicated-kv-store` directory.
Start the leader:

```bash
go run ./cmd/leader -addr :9100
```

Start a follower in another terminal:

```bash
go run ./cmd/follower -addr :9101 -leader http://localhost:9100
```

Write to the leader:

```powershell
Invoke-RestMethod -Method Put `
  -Uri "http://localhost:9100/put?key=language&value=Go"
```

Read from the follower:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:9101/get?key=language"
```

## Try follower catch-up

This procedure shows how a restarted follower recovers from the leader history.

Stop the follower with `Ctrl+C`.
Keep the leader process active.
Write a new key:

```powershell
Invoke-RestMethod -Method Put `
  -Uri "http://localhost:9100/put?key=database&value=systems"
```

The leader returns `sequence=2`. The stopped follower does not have this record.
Start the follower again:

```powershell
go run ./cmd/follower -addr :9101 -leader http://localhost:9100
```

After a successful replication request, read from the follower:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:9101/get?key=database"
```

The expected response is:

```text
found sequence value
True  2        systems
```

The restarted follower starts with an empty map and sequence zero. It requests records after zero. It applies records one and two from the active leader to recover its state.

## Useful endpoints

The leader provides these endpoints:

- `PUT /put?key=...&value=...` to create or update a key.
- `DELETE /delete?key=...` to append a delete record.
- `GET /get?key=...` to read leader state.
- `GET /records?from=N` to return records after sequence `N`.
- `GET /status` to show the leader's latest sequence.

The follower provides these endpoints:

- `GET /get?key=...` to read its local copy.
- `GET /status` to show how far replication has progressed.

This implementation has one leader and keeps state in memory. It does not provide durable storage, automatic failover, or consensus.
