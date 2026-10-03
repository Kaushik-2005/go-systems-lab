# Replicated Key-Value Store

## What it is

This is a small leader/follower key-value store. Clients write to the leader, and follower processes pull the leader's ordered records and apply them locally.

```text
client
  |
  v
leader :9100  -- ordered records -->  follower :9101
                                      follower :9102
```

The followers are intentionally simple. They poll for records after their last sequence number. If a follower goes offline, the leader keeps its log in memory, and the follower catches up when it returns.

## How it works

There are two different roles:

- The **leader** accepts `PUT` and `DELETE` operations.
- A **follower** accepts reads and copies the leader's changes.

The follower does not receive writes directly. It repeatedly asks the leader:

```text
follower has sequence 1
        |
        v
GET /records?from=1
        |
        v
leader returns records 2, 3, ...
        |
        v
follower applies them in order
```

Every leader write creates a record with an increasing sequence number. For example:

```text
sequence  key       value       deleted
1         language  Go          false
2         database  systems     false
3         language              true
```

The sequence is the position of the change in the replication log. It is not a version number for one key. A follower uses it to know exactly where it stopped.

The leader applies a record locally before making it available through `/records`. The follower applies records in order and rejects a gap. Reapplying an old record is safe, so a polling retry does not corrupt the follower.

This is asynchronous replication. A successful leader write does not wait for followers to apply the record. Therefore, a follower can briefly return an older value.

## What a normal write looks like

When this command runs:

```powershell
Invoke-RestMethod -Method Put `
  -Uri "http://localhost:9100/put?key=language&value=Go"
```

the flow is:

1. The leader receives the request.
2. It creates record 1 for `language=Go`.
3. It stores the value locally.
4. It returns the record to the client.
5. The follower asks for records after its current sequence.
6. The follower applies record 1.

The leader response looks like this:

```text
sequence key      value deleted
1        language Go    False
```

The follower response looks like this:

```text
found sequence value
True  1        Go
```

`found=True` means the follower has the key, `value=Go` is the replicated value, and `sequence=1` shows that the follower has applied the first record.

This is asynchronous replication. A successful leader write does not wait for followers to apply the record.

## Project structure

```text
12-replicated-kv-store/
├── store/store.go          # state, ordered records, and replication apply logic
└── cmd/
    ├── leader/main.go      # write API and replication log API
    └── follower/main.go    # polling replica and local read API
```

## Run it

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

This experiment shows what happens when a follower is temporarily unavailable.

First, stop the follower with `Ctrl+C`. Keep the leader running. Then write a new key:

```powershell
Invoke-RestMethod -Method Put `
  -Uri "http://localhost:9100/put?key=database&value=systems"
```

The leader returns `sequence=2`, but the stopped follower has not seen it yet. Start the follower again:

```powershell
go run ./cmd/follower -addr :9101 -leader http://localhost:9100
```

After one polling interval, read from the follower:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:9101/get?key=database"
```

The follower should return:

```text
found sequence value
True  2        systems
```

The follower remembered that its last sequence was 1. After restarting, it requested records after 1, received record 2, and caught up.

## Useful endpoints

The leader exposes:

- `PUT /put?key=...&value=...` to create or update a key;
- `DELETE /delete?key=...` to append a delete record;
- `GET /get?key=...` to read leader state;
- `GET /records?from=N` to return records after sequence `N`;
- `GET /status` to show the leader's latest sequence.

The follower exposes:

- `GET /get?key=...` to read its local copy;
- `GET /status` to show how far replication has progressed.

The implementation keeps state in memory and has one leader. It demonstrates ordered asynchronous replication and reconnect catch-up, not durable storage, automatic failover, or consensus.
