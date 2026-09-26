# Message Queue

## Start with the idea

Imagine a restaurant kitchen. The waiter takes orders quickly, but the kitchen may need more time to prepare them. A queue gives the orders a place to wait:

```text
producer                 broker                  consumer
creates work  ------>  stores work  ------>  processes work
```

The producer and consumer are decoupled. They do not need to run at the same speed, or even be alive at the same time.

That is the basic idea behind a message queue: hold work temporarily and let consumers process it later.

## Queue versus pub/sub

This component is about sharing work between competing consumers:

```text
message-1  --->  consumer A
message-2  --->  consumer B
```

One message is normally handled by one consumer. In pub/sub, the same event is copied to every subscriber. That distinction is why the repository keeps the two components separate.

## What happens to one message?

```text
Publish
   |
   v
pending FIFO --Consume--> in-flight --Ack--> done
                              |
                         no ACK in time
                              |
                              v
                  retry or dead-letter queue
```

The important part is `in-flight`. Consuming a message only means the broker handed it to a consumer. The broker waits for an ACK before treating the work as complete.

## Concepts implemented

- FIFO ordering
- Concurrent producers and consumers
- Message IDs
- Acknowledgements
- In-flight message tracking
- Visibility timeout and redelivery
- Retry limit
- Dead-letter queue

The broker provides at-least-once delivery. A message can be delivered again if its first consumer crashes or takes too long. It does not provide exactly-once processing.

## Project structure

```text
message/
  message.go          message ID, body, and attempt count
queue/
  queue.go            mutex-protected FIFO storage
broker/
  broker.go           publishing, delivery, ACKs, and retries
cmd/
  fifo/               ordering demonstration
  ack/                ACK, timeout, retry, and DLQ demonstration
  concurrent/         multiple producers and consumers
go.mod
design.md             state transitions and failure reasoning
```

## Run it

Run these commands from the `05-message-queue` directory:

```powershell
go run ./cmd/fifo
go run ./cmd/ack
go run ./cmd/concurrent
```

### FIFO example

```text
published: message-1, message-2, message-3
consumed: message-1, message-2, message-3
```

The demo publishes three messages in order and consumes them from the front of the queue. The output shows that the broker preserves FIFO order.

### ACK and failure example

```text
published: pending=3 in_flight=0
consumed: message-1, message-2, message-3 pending=0 in_flight=3
ack message-1: success=true in_flight=2
ack message-3: success=true in_flight=1
unacked message remains: id=message-2 in_flight=1
after first timeout: redelivered=true id=message-2 attempts=2 pending=0 in_flight=1
after retry limit: dead_lettered=true id=message-2 attempts=2 dead_letters=1
```

Read the output as a story:

1. Three messages are published and waiting.
2. All three are consumed, so they move to `in_flight`.
3. Messages 1 and 3 are acknowledged and removed from `in_flight`.
4. Message 2 is intentionally left unacknowledged.
5. Its visibility timeout expires, so the broker delivers it again with attempt 2.
6. It fails again and reaches the retry limit, so it moves to the dead-letter queue.

The code behind this flow is in `broker/broker.go`: `Consume`, `Ack`, `RequeueExpired`, and `StartVisibilityMonitor`.

### Concurrent example

```text
producers=3 consumers=2 produced=30 consumed=30 pending=0 in_flight=0
```

Three producers create thirty messages altogether. Two consumers process and acknowledge them. The final zero counts show that no messages are left pending or in-flight.

## What this version does not do

This is an in-memory, single-process learning implementation. Messages disappear when the broker stops. There is no network protocol, persistence, replay after restart, or transaction connecting consumer work to the ACK.
