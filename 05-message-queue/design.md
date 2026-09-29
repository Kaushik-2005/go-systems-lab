# Message Queue Design

## The problem

Producers and consumers should not need to run at the same speed. The broker stores work in a FIFO queue and gives it to consumers later.

```text
producer -> pending FIFO -> consumer
              ^               |
              |               |
              +--- timeout ---+
```

## Message lifecycle

Messages move through three logical states:

```text
pending -> in-flight -> acknowledged
              |
              +-> timeout -> pending or dead-letter queue
```

`Publish` creates an ID and puts the message in pending. `Consume` removes the oldest message, increments its attempt count, and records a deadline in the in-flight map. `Ack` removes the in-flight entry.

If the consumer crashes or takes too long, the visibility monitor requeues the message. When the attempt count reaches the retry limit, the message moves to the dead-letter queue instead.

This is at-least-once delivery. A message can be delivered more than once, so consumers need idempotent work when duplicates matter.

## Concurrency and repository design

The broker mutex protects IDs, the in-flight map, deadlines, and state transitions. Each FIFO queue protects its own slice. The visibility monitor uses a ticker and stops through a context.

```text
message/           ID, body, and attempts
queue/             FIFO storage
broker/            publish, consume, ACK, retry, and DLQ
cmd/fifo/          ordering demo
cmd/ack/           failure demo
cmd/concurrent/    concurrent producer/consumer demo
```

The broker is in memory and runs in one process. It does not provide durable storage, exactly-once processing, or a transaction joining consumer work with the ACK.
