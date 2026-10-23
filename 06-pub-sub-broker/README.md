# Pub/Sub Broker

## What it is

Let's say one publisher sends an event and multiple independent subscribers need to receive it. A pub/sub broker sends a copy of the event to every subscriber of a topic.

## How it works

The broker maps each topic to its subscription IDs. Every subscription owns a private buffered channel. Unlike a queue, where consumers share work, pub/sub sends each current subscriber its own event copy.

### Subscribe and fan out

![Publishing event E1 to news places a separate copy in subscriber A and subscriber B buffers](diagrams/topic-fanout.svg)

`Subscribe(topic, buffer)` creates a unique subscription ID, allocates its channel, and adds it to the topic. Buffer sizes below one are raised to one. Consumers read from the receive-only channel returned by `Messages()`.

`Publish(topic, event)` snapshots the current subscribers under the broker's read lock, releases that lock, then visits each subscription and sends the event. Reading E1 from subscriber A does not remove it from B. Publishing to an empty topic delivers zero copies, and subscribers joining later receive no replay.

```go
news := broker.New()
reader := news.Subscribe("news", 4)
news.Publish("news", broker.Event{ID: "event-1", Body: "hello"})
event := <-reader.Messages()
```

### Slow-subscriber policies

![Blocking Publish waits for a full channel to drain; PublishNonBlocking skips it and counts a dropped event](diagrams/slow-subscriber-policies.svg)

With capacity one, publishing E1 fills an unread subscriber's channel. When E2 arrives, the policy determines what happens:

| Method | Full subscriber buffer | Result |
| --- | --- | --- |
| `Publish` | Waits until space is available | Returns the number of successful sends |
| `PublishNonBlocking` | Skips that subscriber | Returns delivered and dropped counts |

Blocking publication preserves the send to an open subscription but lets a slow reader delay the publisher. Since subscribers are visited sequentially, it also delays subscribers visited later. Map iteration does not guarantee which subscriber is visited first.

Nonblocking publication uses a channel `select` with a `default` branch, so a full channel does not stall that send. The skipped event is lost for that subscriber. A subscription closed since the snapshot is also counted as a dropped delivery. Both methods still acquire subscription mutexes; nonblocking refers to channel-capacity handling.

### Unsubscribe and channel ownership

![Unsubscribe removes membership before closing the channel; a blocked publish can delay closure until the reader frees buffer space](diagrams/unsubscribe-and-channel-ownership.svg)

The broker owns channel creation, sends, and closure. Consumers receive through `Messages()`, which returns a receive-only channel; they do not close it. Closing a channel signals that no more events will be sent, but does not discard events already in its buffer.

`Unsubscribe(subscription)` first removes membership under the broker lock, removing an empty topic if necessary. It then closes the subscription's channel under the subscription mutex. Calling unsubscribe again returns `false`. Already-buffered events remain available to drain after closure.

The subscription mutex prevents sending on a closed channel. A blocking send holds this mutex while waiting for buffer space, so unsubscribe cannot cancel that send: it waits for the send to finish before closing the channel. A publisher that already took its subscriber snapshot can race with unsubscribe; removal does not instantly cancel that publication.

In the illustrated case, E1 fills a one-event buffer and a publisher blocks sending E2. Unsubscribe removes membership, then waits for the publisher's subscription lock. Reading E1 frees space for E2; the publisher finishes and releases the lock, allowing unsubscribe to close the channel. The reader can still receive E2, followed by `ok == false` once the closed channel is empty. A `for event := range subscription.Messages()` loop drains those buffered events and then ends.

## Concurrency, cost, and limits

The broker's `sync.RWMutex` protects topic membership, and each subscription's mutex protects delivery and closure. Releasing the broker lock before sending lets unrelated topics continue to subscribe and unsubscribe even when one channel is full.

Fan-out visits s subscribers in O(s) work, excluding time waiting for locks or channel capacity. Concurrent publishers can interleave differently for different subscribers; there is no single global order across publishers.

This is an in-memory, single-process broker. It does not provide persistence, replay for late subscribers, acknowledgements, retries, or durable delivery.

## Project structure

```text
broker/             topics, subscriptions, and delivery
cmd/demo/           fan-out and unsubscribe demonstration
cmd/concurrent/     concurrent publisher demonstration
diagrams/           SVG illustrations
go.mod
```

## Run it

Run from the `06-pub-sub-broker` directory:

```powershell
go run ./cmd/demo
go run ./cmd/concurrent
```

The demos show that the first subscribers all receive the same events, unsubscribed subscribers receive nothing new, and concurrent publishers still fan out events to every subscriber.
