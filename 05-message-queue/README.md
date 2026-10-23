# Message Queue

## What it is

Let's say a producer creates work faster than a consumer can process it. A message queue stores the work until a consumer is ready.

## How it works

The broker owns a pending FIFO, an in-flight map, and a dead-letter FIFO. Each message has an ID, body, and attempt count. Consumers compete for pending work: a delivery goes to one consumer at a time.

### FIFO delivery

![Consuming M1 removes the oldest pending message and leaves M2 and M3 for later consumers](diagrams/fifo-delivery.svg)

`Publish(body)` allocates an ID such as `message-1`, initializes the attempt count to zero, and enqueues the message. `Consume()` dequeues the oldest pending message. If none is available, it returns `false` immediately; it does not wait for a producer.

For pending messages `[M1, M2, M3]`, the first consume takes M1 and the next takes M2. Concurrent consumers can finish in a different order. Concurrent publishers can also enqueue in a different order from ID allocation, so message IDs do not define FIFO order.

The FIFO uses a mutex-protected slice. Enqueue appends to the tail; dequeue clears the first slot and advances the slice. Enqueue is amortized O(1), and dequeue is O(1).

### In-flight messages and acknowledgements

![Messages move from pending to in-flight, then ACK removes them or expiration sends them to retry or dead-letter storage](diagrams/message-lifecycle.svg)

On consume, the broker increments the message's attempt count and records `deadline = now + visibilityTimeout` in the in-flight map. The message is no longer pending, so another consumer cannot immediately take it.

After processing succeeds, the consumer calls `Ack(id)`. An ACK removes the in-flight entry and returns `true`; an ID not currently in flight returns `false`. The broker does not retain a permanent history of completed messages.

### Visibility timeout, retries, and dead letters

![With retryLimit two, the first timeout requeues M1 and the second unacknowledged delivery times out into the dead-letter queue](diagrams/visibility-and-retries.svg)

`RequeueExpired()` removes deliveries whose deadlines have arrived. If `Attempts < retryLimit`, it appends the message to the pending tail. Otherwise it sends the message to the dead-letter queue. The message keeps its ID and body across deliveries.

In this implementation, `retryLimit = 2` means two total delivery attempts: the initial delivery and one redelivery. If neither is acknowledged, the second expiration moves the message to dead-letter storage. `ConsumeDeadLetter()` lets the caller inspect and remove those messages.

`StartVisibilityMonitor(ctx, interval)` runs expiration scans on a ticker until the context is canceled. Expiration is observed during a scan, so redelivery can occur later than the exact deadline. Without the monitor or explicit calls to `RequeueExpired`, expired messages stay in flight.

### Delivery guarantees

Unacknowledged work can be delivered again. For example, a consumer may complete its side effect and crash before its ACK reaches the broker. Consumers should make processing idempotent when duplicates matter.

ACKs identify the message, rather than a specific delivery attempt. A delayed ACK from an earlier attempt can therefore remove a newer in-flight delivery of the same message. This is a simplified at-least-once model with bounded retries, rather than exactly-once processing.

## Concurrency and limits

The broker mutex protects IDs and the in-flight map. Each FIFO protects its own slice. Moving a message between queues and the in-flight map involves separate lock acquisitions; these steps are not one transaction, and length methods do not provide an atomic snapshot of all three states.

An expiration scan is O(i) for i in-flight messages. Pending and dead-letter queues are unbounded. Everything is in memory in one process, so a restart loses all messages. There is no durable storage or transaction joining consumer work with the ACK.

## Project structure

```text
message/           message identity and payload
queue/             FIFO storage
broker/            publishing, consuming, ACKs, and retries
cmd/fifo/          ordering demonstration
cmd/ack/           failure and retry demonstration
cmd/concurrent/    concurrent producers and consumers
diagrams/          SVG illustrations
go.mod
```

## Run it

Run from the `05-message-queue` directory:

```powershell
go run ./cmd/fifo
go run ./cmd/ack
go run ./cmd/concurrent
```

The demos show FIFO order, acknowledgement state, redelivery after timeout, dead-letter handling, and concurrent processing.
