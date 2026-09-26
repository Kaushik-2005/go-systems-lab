# Message Queue Design

## The question this component answers

What should happen when a producer creates work faster than a consumer can process it?

The queue gives that work somewhere to wait:

```text
producer --Publish--> pending FIFO --Consume--> consumer
                         ^                         |
                         |                         |
                         +------ timeout/no ACK ---+
```

The interesting part is not just storing messages. It is deciding when a message is still being worked on, when it should be tried again, and when it has failed too many times.

## The three states

Every message is in one of these logical states:

```text
pending       waiting to be delivered
in-flight     delivered, but not acknowledged
dead-letter   failed enough times to stop retrying
```

The broker stores them as:

```text
pending   *queue.Queue
inflight  map[string]inFlightMessage
dead      *queue.Queue
```

`queue.Queue` owns FIFO storage. `Broker` owns IDs, deadlines, retry decisions, and transitions between states.

## Message model

```text
ID        broker-generated identifier, for example message-2
Body      the work or event payload
Attempts  number of times the broker has delivered it
```

The ID is important because the consumer uses it to acknowledge a specific delivery. The attempt count tells the broker whether to retry or dead-letter an expired message.

## How a message moves through the broker

### 1. Publish

`Publish` creates an ID and appends the message to the pending FIFO.

```text
Publish("send email")
        |
        v
pending: [message-1]
```

The producer does not wait for the work to finish. That separation is the main reason to use a queue.

### 2. Consume

`Consume` removes the oldest pending message, increments `Attempts`, and records it in the in-flight map with a deadline.

```text
pending --Consume--> in-flight(message ID, attempts, deadline)
```

The message is no longer available in pending, so two consumers cannot remove the same pending entry at once.

### 3. ACK

After processing succeeds, the consumer calls `Ack(id)`.

```text
in-flight --Ack(message-1)--> completed
```

The broker removes the in-flight entry. It does not need a completed-message store for this learning implementation.

## Why ACKs exist

Suppose the broker removed a message as soon as it delivered it:

```text
1. broker delivers message-2
2. consumer crashes
3. message-2 is gone forever
```

Keeping it in-flight until an ACK avoids that loss. But keeping it there forever would also be a problem if the consumer disappears. The visibility timeout solves both problems.

## Visibility timeout and redelivery

Each delivery receives a deadline. A background monitor periodically calls `RequeueExpired`:

```text
Consume
   |
   v
in-flight + deadline
   |
   +-- ACK before deadline --> completed
   |
   +-- deadline passes ------> pending again
```

`StartVisibilityMonitor` runs this check on a ticker and stops when its context is cancelled.

This gives at-least-once delivery, not exactly-once delivery. A slow consumer might still be working when its deadline expires. The broker can then redeliver the message to another consumer, so duplicate processing is possible.

## Retry limit and dead letters

When an expired message is examined, the broker chooses one path:

```text
attempts < retry limit  -> pending for another try
attempts >= retry limit -> dead-letter queue
```

The dead-letter queue is a place to inspect messages that keep failing. This implementation does not automatically repair, delete, or republish them.

## Concurrency and ownership

There are two layers of synchronization:

- `Broker.mu` protects message ID generation, the in-flight map, and state transitions.
- `queue.Queue` has its own mutex protecting enqueue, dequeue, and length operations.

The concurrent demo starts three producers and two consumers. Producers can publish at the same time, and consumers can consume at the same time, without corrupting the queues or the in-flight map.

The visibility monitor is the only long-running background goroutine. Its context controls its lifetime, which prevents it from running after the broker is no longer needed.

## Failure behavior

| Situation | What happens |
|---|---|
| Consumer crashes before ACK | The message stays in-flight until timeout, then is redelivered. |
| Consumer is slower than the timeout | The message may be delivered twice. |
| ACK arrives after timeout | The original in-flight entry may be gone, so ACK can return `false`. |
| Message keeps failing | It moves to the dead-letter queue at the retry limit. |
| Broker process stops | All messages disappear because storage is in memory. |

## What this design guarantees

This version gives us FIFO ordering in the pending queue, synchronized in-process access, explicit ACK tracking, timeout-based redelivery, and bounded retries.

It does not provide durable storage, cross-process communication, exactly-once processing, or a transaction that atomically combines consumer work with an ACK. Those are separate problems for later components.
