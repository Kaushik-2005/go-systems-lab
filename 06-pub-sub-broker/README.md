# Pub/Sub Broker

## Start with the idea

Think of a news channel. A publisher posts one update, and everyone subscribed to that channel gets their own copy:

```text
                 +--> subscriber A
publisher --> news topic
                 +--> subscriber B
```

That is publish/subscribe, usually shortened to pub/sub.

The key difference from a message queue is the delivery rule:

```text
message queue:  one message -> one competing consumer
pub/sub:       one event   -> every subscriber of the topic
```

A queue distributes work. Pub/sub broadcasts an event to independent consumers, such as an audit service, notification service, and search indexer.

## Event lifecycle

```text
Publish(event)
       |
       v
 topic: news
    /       \\
   v         v
subscriber A  subscriber B
```

Each subscriber has its own channel. Reading the event from A does not remove it from B's channel.

## What is implemented?

- Named topics
- Multiple subscribers per topic
- Fan-out delivery
- Subscribe and unsubscribe
- Subscriber channel cleanup
- Concurrent publication
- Explicit event IDs and bodies
- Two slow-subscriber policies:
  - `Publish`: wait when a subscriber buffer is full
  - `PublishNonBlocking`: drop for a full subscriber and continue

## Project structure

```text
broker/
  broker.go            events, topics, subscriptions, and delivery
cmd/
  demo/
    main.go            fan-out, unsubscribe, and slow-subscriber demo
  concurrent/
    main.go            concurrent publisher demo
go.mod
design.md              state ownership and delivery reasoning
```

## Run it

Run from the `06-pub-sub-broker` directory:

```powershell
go run ./cmd/demo
go run ./cmd/concurrent
```

## Follow the main demo

The demo prints the subscriber histories after every publication:

```text
published event-1 ("one") to 2 subscribers
  A has: [event-1(one)]
  B has: [event-1(one)]

published event-2 ("two") to 2 subscribers
  A has: [event-1(one), event-2(two)]
  B has: [event-1(one), event-2(two)]

published event-3 ("three") to 2 subscribers
  A has: [event-1(one), event-2(two), event-3(three)]
  B has: [event-1(one), event-2(two), event-3(three)]

unsubscribe A: success=true
published event-4 ("four") to 1 subscribers
  A has: [event-1(one), event-2(two), event-3(three)]
  B has: [event-1(one), event-2(two), event-3(three), event-4(four)]
```

Here is how to read it:

1. Events 1, 2, and 3 appear in both histories. That is fan-out.
2. Unsubscribing A removes it from the topic and closes its channel.
3. Event 4 goes only to B. A keeps its old history but receives nothing new.

The IDs make it easy to follow one publication through both subscribers.

## Slow subscribers

The demo also shows the non-blocking policy:

```text
non-blocking policy: alert-1 delivered=1 dropped=0
non-blocking policy: alert-2 delivered=0 dropped=1
slow subscriber received: [{alert-1 first}]
```

The subscriber has a buffer of one. It accepts `alert-1`, but nobody reads it before `alert-2` arrives. Because the demo uses `PublishNonBlocking`, the second event is dropped instead of blocking the publisher.

This is a real design trade-off: blocking protects delivery but can slow publishers; dropping protects throughput but loses events for slow subscribers.

## Concurrent publication

```text
publishers=3 events_each=5 expected=15
subscriber A received=15
subscriber B received=15
fan-out complete=true
```

Three publisher goroutines create fifteen events. Both subscribers receive all fifteen. The exact order of events from different publishers can vary because those goroutines run concurrently.

## Scope

This is an in-memory, single-process learning implementation. It has no network protocol, persistence, replay for late subscribers, acknowledgements, or durable delivery. It does not claim exactly-once processing.
