# Leader Election Design

## The idea

A cluster needs one coordinator, but a coordinator can fail. Leader election lets the remaining nodes choose a replacement while using terms to reject stale messages.

The implementation has two layers:

- election.Node contains the state machine and voting rules;
- cmd/node exposes those rules over HTTP and runs timers.

## Node state

Each node stores:

- its ID;
- follower, candidate, or leader state;
- current term;
- the candidate it voted for in that term;
- votes received during its current election;
- current leader ID;
- a channel used to reset the election timer.

A mutex protects the state because HTTP handlers and timer goroutines can access it concurrently.

## Election flow

1. A follower waits for a randomized timeout.
2. If no heartbeat arrives, it increments its term and becomes a candidate.
3. It votes for itself and sends vote requests to its peers.
4. Each peer grants at most one vote in that term.
5. The candidate records granted votes.
6. A majority promotes it to leader.
7. The leader sends heartbeats at a fixed interval.
8. Followers reset their election timers when valid heartbeats arrive.
9. If the leader stops, a follower times out and starts a new election.
10. A message with a higher term makes an old leader or candidate step down.

For three nodes, two votes are enough. The candidate's self-vote counts toward the majority.

## HTTP messages

The node process exposes:

- POST /vote for vote requests;
- POST /heartbeat for leader heartbeats;
- GET /status for observable state.

The peer list is supplied when a process starts. Network failures are treated as missing votes or missed heartbeats, which allows the remaining nodes to elect a replacement.

## Terms and stale messages

A node rejects vote requests and heartbeats from older terms. A newer term resets the node's previous election state. This prevents delayed messages from an earlier election from making an old candidate leader.

The randomized timeout reduces the chance that every follower starts an election at exactly the same time. It does not eliminate split votes completely; a later timeout can trigger another election.

## Failure behavior and scope

If the leader stops, followers eventually stop receiving heartbeats and elect a replacement. If fewer than a majority of nodes are reachable, no candidate can become leader.

This is a leader-election mechanism inspired by Raft, not complete Raft consensus. It has no replicated application log, durable terms, quorum-backed writes, or persistent membership.
