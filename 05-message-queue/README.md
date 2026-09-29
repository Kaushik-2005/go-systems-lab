# Message Queue

## What it is

Lets say a producer creates work faster than a consumer can process it. A message queue stores the work until a consumer is ready.

```text
producer -> broker queue -> consumer
```

## How it works

Messages enter a FIFO queue and move to an in-flight state when consumed.

```text
pending -> consume -> in-flight -> ACK -> done
                         |
                         +-> timeout -> retry or dead-letter queue
```

The broker also handles:

- concurrent producers and consumers;
- message IDs;
- acknowledgements;
- visibility timeout and redelivery;
- retry limits;
- dead-letter messages.

The delivery model is at-least-once. An unacknowledged message can be delivered again.

## Project structure

```text
message/           message identity and payload
queue/             FIFO storage
broker/            publishing, consuming, ACKs, and retries
cmd/fifo/          ordering demonstration
cmd/ack/           failure and retry demonstration
cmd/concurrent/    concurrent producers and consumers
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
