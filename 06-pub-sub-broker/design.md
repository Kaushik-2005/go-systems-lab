# Pub/Sub Broker Design

## The problem

One publisher may need to notify several independent consumers. Unlike a queue, where consumers share work, pub/sub gives every subscriber its own copy.

```text
Publish(event) -> topic -> subscriber A
                     \-> subscriber B
```

The broker stores topics as a map of subscription IDs. Each subscription owns a private buffered channel, so reading an event from A does not remove it from B.

## Publish and subscription flow

`Subscribe` creates an ID, allocates a channel, and adds the subscriber to a topic. `Publish` snapshots the subscribers, releases the broker lock, and sends the event to each private channel. `Unsubscribe` removes the subscriber first and then closes its channel.

Releasing the broker lock before channel sends matters because a full buffer can block. It should not prevent unrelated subscribe or unsubscribe operations.

## Slow subscribers and concurrency

`Publish` waits when a subscriber buffer is full. This preserves delivery but can slow the publisher. `PublishNonBlocking` skips a full subscriber and counts the event as dropped. This protects publisher throughput but loses events for that subscriber.

The broker mutex protects topic membership. Each subscription mutex protects its closed state and coordinates delivery with channel closure. Concurrent publishers can interleave events from different goroutines; there is no single global order across publishers.

```text
broker/broker.go       topics, subscriptions, fan-out, policies
cmd/demo/main.go       fan-out and unsubscribe demo
cmd/concurrent/main.go concurrent publisher demo
```

The broker is an in-memory, single-process component. It does not provide persistence, replay for late subscribers, acknowledgements, or durable delivery.
