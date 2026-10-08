# Leader Election

## What it is

Leader election chooses one coordinator from a group of nodes. The other nodes follow that leader and can elect a replacement if the leader disappears.

This component now uses separate local processes that communicate over HTTP.

text
node-a  <---- vote requests / heartbeats ---->  node-b
  ^                                             ^
  +----------------------> node-c -------------+

## How it works

- Each node starts as a follower.
- A follower waits for a randomized election timeout.
- If no heartbeat arrives before the timeout, it becomes a candidate and increases its term.
- The candidate asks its peers for votes.
- Each peer grants at most one vote in a term.
- A candidate with a majority becomes leader.
- The leader sends regular heartbeats.
- If the leader stops, followers time out and elect a new leader in a higher term.
- A node steps down when it sees a higher term.

The status endpoint exposes the node ID, state, term, current leader, and vote count.

## Project structure

text
14-leader-election/
├── election/node.go       # state machine, voting rules, and timers
└── cmd/
    ├── demo/main.go       # in-process timer demonstration
    └── node/main.go       # networked HTTP node

## Run the in-process demo

go run ./cmd/demo

This shows heartbeats preventing an election and a timeout starting one after heartbeats stop.

## Run a three-node cluster

Open three terminals from this directory.

Terminal 1:

go run ./cmd/node -id node-a -addr :9201 -peers localhost:9202,localhost:9203

Terminal 2:

go run ./cmd/node -id node-b -addr :9202 -peers localhost:9201,localhost:9203

Terminal 3:

go run ./cmd/node -id node-c -addr :9203 -peers localhost:9201,localhost:9202

Check each node:

Invoke-RestMethod http://localhost:9201/status
Invoke-RestMethod http://localhost:9202/status
Invoke-RestMethod http://localhost:9203/status

One node should show leader. The other two should show follower and the same term and leader ID.

## Try leader failure

Stop the current leader with Ctrl+C. Wait for the remaining nodes to elect a replacement, then check their status endpoints again.

The new leader should have:

- a higher term;
- state=leader;
- a majority of votes;
- a different ID from the stopped leader.

This is a simplified election service inspired by Raft. It does not replicate application logs or provide complete consensus.
