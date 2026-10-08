package election

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

type State string

const (
	Follower  State = "follower"
	Candidate State = "candidate"
	Leader    State = "leader"
)

type Node struct {
	mu         sync.Mutex
	id         string
	state      State
	term       uint64
	votedFor   string
	leader     string
	votes      map[string]bool
	timerReset chan struct{}
}

func NewNode(id string) *Node {
	return &Node{id: id, state: Follower, votes: make(map[string]bool), timerReset: make(chan struct{}, 1)}
}

func (n *Node) StartElection() uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.term++
	n.state = Candidate
	n.votedFor = n.id
	n.leader = ""
	n.votes = map[string]bool{n.id: true}
	return n.term
}

func (n *Node) HandleVoteRequest(term uint64, candidate string) (bool, uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if term < n.term {
		return false, n.term
	}
	if term > n.term {
		n.term = term
		n.state = Follower
		n.votedFor = ""
		n.leader = ""
		n.votes = make(map[string]bool)
	}
	if n.votedFor != "" && n.votedFor != candidate {
		return false, n.term
	}

	n.votedFor = candidate
	return true, n.term
}

func (n *Node) ReceiveVote(term uint64, voter string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != Candidate || n.term != term || n.votes[voter] {
		return false
	}
	n.votes[voter] = true
	return true
}

func (n *Node) BecomeLeader(clusterSize int) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != Candidate || len(n.votes)*2 <= clusterSize {
		return false
	}
	n.state = Leader
	n.leader = n.id
	return true
}

func (n *Node) HandleHeartbeat(term uint64, leader string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if term < n.term {
		return false
	}
	if term > n.term {
		n.term = term
		n.votedFor = ""
		n.votes = make(map[string]bool)
	}
	n.state = Follower
	n.leader = leader
	n.resetTimerLocked()
	return true
}

func (n *Node) StartElectionTimer(ctx context.Context, minimum, maximum time.Duration, onTimeout func()) error {
	if minimum <= 0 || maximum < minimum {
		return errors.New("invalid election timeout range")
	}

	go func() {
		for {
			timer := time.NewTimer(randomTimeout(minimum, maximum))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				onTimeout()
			case <-n.timerReset:
				if !timer.Stop() {
					<-timer.C
				}
			}
		}
	}()
	return nil
}

func (n *Node) resetTimerLocked() {
	select {
	case n.timerReset <- struct{}{}:
	default:
	}
}

func randomTimeout(minimum, maximum time.Duration) time.Duration {
	if minimum == maximum {
		return minimum
	}
	return minimum + time.Duration(rand.Int63n(int64(maximum-minimum)))
}

type Snapshot struct {
	ID        string
	State     State
	Term      uint64
	VotedFor  string
	Leader    string
	VoteCount int
}

func (n *Node) Snapshot() Snapshot {
	n.mu.Lock()
	defer n.mu.Unlock()

	return Snapshot{ID: n.id, State: n.state, Term: n.term, VotedFor: n.votedFor, Leader: n.leader, VoteCount: len(n.votes)}
}
