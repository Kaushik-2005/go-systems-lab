# Pub/Sub Broker

## What it is

Lets say one publisher sends an event and multiple independent subscribers need to receive it. A pub/sub broker sends a copy of the event to every subscriber of a topic.

```text
publisher -> topic -> subscriber A
                  -> subscriber B
```

## How it works

Subscribers join named topics. When an event is published, the broker sends it to each subscriber's own channel.

```text
Publish(event)
       |
       v
    news topic
     /      \\
    v        v
subscriber A  subscriber B
```

The broker also handles:

- multiple subscribers per topic;
- fan-out delivery;
- subscribe and unsubscribe;
- concurrent publication;
- blocking slow-subscriber handling;
- non-blocking message dropping for full buffers.

## Project structure

```text
broker/             topics, subscriptions, and delivery
cmd/demo/           fan-out and unsubscribe demonstration
cmd/concurrent/     concurrent publisher demonstration
go.mod
```

## Run it

Run from the `06-pub-sub-broker` directory:

```powershell
go run ./cmd/demo
go run ./cmd/concurrent
```

The demos show that the first subscribers all receive the same events, unsubscribed subscribers receive nothing new, and concurrent publishers still fan out events to every subscriber.
