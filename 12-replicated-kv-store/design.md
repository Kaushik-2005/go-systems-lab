# Replicated Key-Value Store Design

## The idea

A single key-value store is simple to use but a process failure makes its state unavailable. Replication keeps copies of the state in other processes.

The first version uses one leader and followers. The leader is the only writer. Followers do not invent their own updates; they apply the leader's ordered records.

## Leader state and replication flow

The store holds a map of current values and an in-memory log of records:

```text
sequence  key       value       deleted
1         language  Go          false
2         cache                 true
```

Every `PUT` or `DELETE` increments the sequence and appends a record. The follower tracks only its last applied sequence and asks the leader for records after it.

```text
follower sequence = 2
          |
          v
GET /records?from=2
          |
          v
apply records 3, 4, 5 in order
```

`Apply` rejects a gap. For example, record 5 cannot be applied when record 4 is missing. This keeps the follower's state ordered instead of silently applying updates out of sequence. Reapplying an old record is harmless, which makes polling retries idempotent.

## Concurrency

The store uses one `sync.RWMutex`. Writes and replication apply operations take the write lock because they update both the map and sequence. Reads and log copies take the read lock.

The follower's polling goroutine can run while HTTP readers access the local store. The mutex protects that shared state. A slow or unavailable leader only causes a failed polling attempt; the follower keeps serving its last known state.

## Why asynchronous replication

The leader responds after applying its own write and does not wait for followers. This keeps the write path simple and makes replication lag visible. The trade-off is that a follower can return stale data, and a leader failure can lose writes that had not reached a follower.

Synchronous replication would wait for one or more followers before acknowledging a write, improving durability at the cost of write latency and availability. That is a useful next variation, but it is deliberately not hidden in this first version.

## Failure behavior

If a follower stops, the leader's in-memory log retains records. When the follower restarts, it asks for records after its last sequence and catches up.

If the leader stops, followers keep serving their last known values, but no new writes can be accepted. There is no automatic leader election, so another process cannot safely take over.

If records arrive out of order, the follower rejects them. If the same record is received twice, the sequence check treats it as already applied. Since the log and state are in memory, restarting the leader or a follower loses that process's state.
