package main

import (
	"context"
	"fmt"
	"time"

	"leader-election/election"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	follower := election.NewNode("node-b")
	timeout := make(chan struct{}, 1)

	err := follower.StartElectionTimer(ctx, 100*time.Millisecond, 180*time.Millisecond, func() {
		select {
		case timeout <- struct{}{}:
		default:
		}
		follower.StartElection()
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("sending heartbeats for 500ms")
	heartbeatDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		defer close(heartbeatDone)

		deadline := time.After(500 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				follower.HandleHeartbeat(1, "node-a")
			case <-deadline:
				return
			}
		}
	}()

	<-heartbeatDone
	fmt.Println("heartbeats stopped; waiting for election timeout")

	select {
	case <-timeout:
		state := follower.Snapshot()
		fmt.Printf("timeout triggered: state=%s term=%d\n", state.State, state.Term)
	case <-time.After(time.Second):
		fmt.Println("timeout did not trigger")
	}
}
