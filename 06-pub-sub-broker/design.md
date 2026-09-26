# Pub/Sub Broker Design

## The question this component answers

How can one publisher announce an event to many independent consumers?

The broker uses a topic as the meeting point:

```text
Publish("news", event)
          |
          v
       news topic
       /        \\
      v          v
 subscriber A  subscriber B
```

Both subscribers receive a copy. This is the important difference from a message queue, where competing consumers share messages between them.

## The state model

The broker stores topic membership like this:

```text
topics[topic][subscriptionID] = subscription
```

In Go, that is:

```text
map[string]map[int]*Subscription
```

Each `Subscription` owns its own buffered channel. A topic does not have one shared channel because reading a shared channel would remove the event for everyone else.

An `Event` contains:

```text
ID    publication identifier, such as event-2
Body  the event payload
```

The ID is not required for routing, but it makes the demo easier to follow.

## Publishing step by step

When `Publish` is called:

1. The broker takes a read lock.
2. It copies the current subscribers for the topic into a slice.
3. It releases the broker lock.
4. It sends the event to each subscriber's private channel.

```text
publisher -> topic subscribers -> A's channel
                              \-> B's channel
```

The broker lock is released before channel sends. That matters because a channel send may wait for buffer space. Holding the broker lock while waiting would unnecessarily block subscribe and unsubscribe operations for the whole broker.

## Subscribing and unsubscribing

`Subscribe` creates a unique subscription ID, allocates a buffered channel, and adds it to the topic map.

`Unsubscribe` does two things in order:

1. Removes the subscription from the topic map.
2. Closes its channel.

Removing it first means new publications will not select it. The subscription mutex coordinates delivery with channel closure so a send does not race with `close`.

## Slow subscribers: two deliberate policies

Every subscriber has a bounded channel buffer. Eventually a subscriber can stop reading and fill that buffer. There is no universally correct answer for what should happen next, so the broker makes the choice explicit.

### Blocking policy

`Publish` waits when a subscriber's buffer is full:

```text
buffer has room -> deliver and continue
buffer is full  -> wait
```

This avoids dropping events, but a slow subscriber can slow down the publisher.

### Drop policy

`PublishNonBlocking` uses a non-blocking channel send:

```text
buffer has room -> deliver
buffer is full  -> count as dropped and continue
```

This protects publisher throughput but sacrifices delivery for slow subscribers. The demo creates a buffer of one, accepts `alert-1`, and drops `alert-2` because the first alert has not been read yet.

Other possible policies include disconnecting slow subscribers, increasing the buffer, or giving each subscriber a dedicated delivery goroutine. The learning goal is to see the trade-off before hiding it behind an abstraction.

## Concurrent publishers

The concurrent demo starts three publisher goroutines and two subscribers:

```text
publisher 1 --\\
publisher 2 ----> broker ---> subscriber A
publisher 3 --/          \\--> subscriber B
```

Both subscribers receive all fifteen events. However, there is no single global order across publishers. Events from one publisher follow that publisher's loop order, while events from different publishers can be interleaved by the scheduler.

## Concurrency model

- `Broker.mu` protects the topic map and subscription membership.
- `Subscription.mu` protects its `closed` flag and coordinates sends with channel closure.
- Channels transfer events from publishers to subscribers.
- The broker owns topic membership; each subscription owns its channel lifecycle.

The implementation is safe for concurrent subscribe, publish, and unsubscribe calls. It is still an in-memory, single-process broker; concurrency safety does not make it a distributed service.

## Queue and pub/sub semantics

```text
Message queue:
  publish one message
      |
      +--> one competing consumer handles it

Pub/sub:
  publish one event
      |
      +--> subscriber A receives a copy
      +--> subscriber B receives a copy
```

Use a queue when consumers are sharing work. Use pub/sub when independent consumers must each see the event.

## Failure behavior and scope

| Situation | Result |
|---|---|
| Subscriber stops reading with `Publish` | Its full buffer can block the publisher. |
| Subscriber stops reading with `PublishNonBlocking` | Events for that subscriber can be dropped. |
| Subscriber unsubscribes | It is removed from the topic and its channel closes. |
| Broker process stops | All topics and buffered events disappear. |
| Subscriber joins late | It receives only future events; there is no replay. |

This design does not provide persistence, network communication, acknowledgements, replay, durable subscriptions, or exactly-once processing.
