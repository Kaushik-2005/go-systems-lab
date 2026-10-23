# Leader Election

## What it is

Leader election chooses a coordinator from a fixed group of nodes. This component has a voting state machine and a small HTTP node process that exchanges vote requests and heartbeats with configured peers.

## How it works

### Follower, candidate, and leader states

![Timeout starts a new candidacy, majority votes elect a leader, and newer terms cause step-down](diagrams/election-states.svg)

`NewNode(id)` starts as a follower. `StartElection` increments its term, clears the known leader, becomes a candidate, and records its self-vote. The node process requests votes from peers and records successful responses for that term.

Each node stores its ID, role, term, candidate voted for, current leader ID, received voter IDs, and a buffered timer-reset channel. A mutex protects this state.

### Terms and majority voting

![A candidate self-vote and one peer vote form a two-of-three majority even when the third node is unreachable](diagrams/majority-voting.svg)

`HandleVoteRequest` rejects an older term. A higher term makes the receiver a follower and clears its previous vote. Within one term, a node can grant repeated requests for the same candidate but cannot vote for a different one. `ReceiveVote` ignores duplicate voters and votes that do not match the active candidacy's term.

`BecomeLeader(clusterSize)` requires more than half the cluster's votes, including the self-vote. In a three-node group, two votes suffice. Network failures count as missing responses; an isolated node cannot win with only its self-vote.

The HTTP command supplies the cluster size from its configured peer list. Membership is static; the state machine assumes supplied identities and peer configuration are valid.

### Heartbeats, timeouts, and failover

![After a leader stops, a follower times out and can become the next leader with a reachable majority](diagrams/heartbeat-failover.svg)

The command uses randomized timeouts from 300 ms to under 600 ms. An accepted heartbeat resets the follower's timer. The leader's heartbeat loop ticks every 100 ms. Randomization reduces simultaneous candidacies; a failed or split election can retry after another timeout.

Older-term heartbeats are rejected. A higher-term heartbeat updates the term and clears the old vote; an accepted heartbeat makes the receiver a follower and records the leader, including when a candidate receives a heartbeat in its current term. Higher-term vote responses and heartbeat responses also make the command step down.

If the leader stops, reachable followers stop receiving heartbeats and can elect a replacement in a higher term. Fewer than a majority of reachable nodes cannot elect a leader. Heartbeats are sent sequentially to peers with a 500 ms client timeout, so an unavailable peer can delay that loop.

## HTTP messages

| Request | Purpose |
| --- | --- |
| `POST /vote` | Candidate asks for a vote using its term and ID |
| `POST /heartbeat` | Leader announces its term and ID |
| `GET /status` | Inspect ID, role, term, leader, and vote count |

## Concurrency and limits

HTTP handlers and timer goroutines share the node mutex. The command also serializes its own election attempts, while sending vote requests to peers concurrently. Each election contacts O(p) peers for p configured peers.

This is inspired by Raft election rules, but does not replicate an application log or implement complete consensus. Terms and votes are not durable, membership is not persisted, and leadership alone does not provide quorum-backed writes or fencing for application work.

## Project structure

```text
election/node.go    state machine, voting rules, and timers
cmd/demo/main.go    in-process heartbeat / timeout demo
cmd/node/main.go    networked HTTP node
diagrams/           SVG illustrations
go.mod
```

## Run the in-process demo

Run from the `14-leader-election` directory:

```powershell
go run ./cmd/demo
```

This sends heartbeats for 500 ms, then shows a timeout starting an election after heartbeats stop.

## Run a three-node cluster

Open three terminals in this component directory and run one command in each:

```powershell
go run ./cmd/node -id node-a -addr :9201 -peers localhost:9202,localhost:9203
```

```powershell
go run ./cmd/node -id node-b -addr :9202 -peers localhost:9201,localhost:9203
```

```powershell
go run ./cmd/node -id node-c -addr :9203 -peers localhost:9201,localhost:9202
```

Inspect the nodes:

```powershell
Invoke-RestMethod http://localhost:9201/status
Invoke-RestMethod http://localhost:9202/status
Invoke-RestMethod http://localhost:9203/status
```

Once the cluster settles, one node should report leader and the others follower with the same leader and term. Stop the current leader with `Ctrl+C`, wait for a replacement election, and inspect the remaining nodes again. The replacement should have a higher term and a majority of votes.
