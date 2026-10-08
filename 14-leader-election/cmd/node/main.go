package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"leader-election/election"
)

type service struct {
	node       *election.Node
	peers      []string
	client     *http.Client
	electionMu sync.Mutex
}

type voteRequest struct {
	Term      uint64
	Candidate string
}

type voteResponse struct {
	Granted bool
	Term    uint64
}

type voteResult struct {
	response voteResponse
	voter    string
}

type heartbeatRequest struct {
	Term   uint64
	Leader string
}

func main() {
	id := flag.String("id", "", "node ID")
	address := flag.String("addr", ":9201", "HTTP listen address")
	peerList := flag.String("peers", "", "comma-separated peer URLs")
	flag.Parse()

	if *id == "" {
		log.Fatal("-id is required")
	}

	s := &service{
		node:   election.NewNode(*id),
		peers:  splitPeers(*peerList),
		client: &http.Client{Timeout: 500 * time.Millisecond},
	}

	http.HandleFunc("/vote", s.vote)
	http.HandleFunc("/heartbeat", s.heartbeat)
	http.HandleFunc("/status", s.status)

	ctx := context.Background()
	if err := s.node.StartElectionTimer(ctx, 300*time.Millisecond, 600*time.Millisecond, func() {
		go s.startElection()
	}); err != nil {
		log.Fatal(err)
	}
	go s.sendHeartbeats(ctx)

	log.Printf("%s listening on %s with peers %v", *id, *address, s.peers)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func (s *service) vote(w http.ResponseWriter, r *http.Request) {
	var request voteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid vote request", http.StatusBadRequest)
		return
	}

	granted, term := s.node.HandleVoteRequest(request.Term, request.Candidate)
	writeJSON(w, voteResponse{Granted: granted, Term: term})
}

func (s *service) heartbeat(w http.ResponseWriter, r *http.Request) {
	var request heartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid heartbeat", http.StatusBadRequest)
		return
	}

	accepted := s.node.HandleHeartbeat(request.Term, request.Leader)
	writeJSON(w, map[string]any{"accepted": accepted, "term": s.node.Snapshot().Term})
}

func (s *service) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.node.Snapshot())
}

func (s *service) startElection() {
	s.electionMu.Lock()
	defer s.electionMu.Unlock()

	before := s.node.Snapshot()
	if before.State == election.Leader {
		return
	}

	term := s.node.StartElection()
	votes := 1
	results := make(chan voteResult, len(s.peers))
	var waitGroup sync.WaitGroup

	for _, peer := range s.peers {
		waitGroup.Add(1)
		go func(peer string) {
			defer waitGroup.Done()
			var response voteResponse
			if err := s.post(peer+"/vote", voteRequest{Term: term, Candidate: before.ID}, &response); err == nil {
				results <- voteResult{response: response, voter: peer}
			}
		}(peer)
	}

	waitGroup.Wait()
	close(results)

	for result := range results {
		response := result.response
		if response.Term > term {
			s.node.HandleHeartbeat(response.Term, "")
			return
		}
		if response.Granted {
			votes++
			s.node.ReceiveVote(term, result.voter)
		}
	}

	if s.node.BecomeLeader(len(s.peers) + 1) {
		log.Printf("%s became leader for term %d with %d votes", before.ID, term, votes)
	}
}

func (s *service) sendHeartbeats(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			snapshot := s.node.Snapshot()
			if snapshot.State != election.Leader {
				continue
			}
			for _, peer := range s.peers {
				var response map[string]any
				if err := s.post(peer+"/heartbeat", heartbeatRequest{Term: snapshot.Term, Leader: snapshot.ID}, &response); err != nil {
					continue
				}
				if term, ok := response["term"].(float64); ok && uint64(term) > snapshot.Term {
					s.node.HandleHeartbeat(uint64(term), "")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *service) post(endpoint string, request any, response any) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	httpResponse, err := s.client.Post(endpoint, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("peer returned %s", httpResponse.Status)
	}
	return json.NewDecoder(httpResponse.Body).Decode(response)
}

func splitPeers(value string) []string {
	var peers []string
	for _, peer := range strings.Split(value, ",") {
		peer = strings.TrimSpace(peer)
		if peer == "" {
			continue
		}
		if !strings.HasPrefix(peer, "http://") && !strings.HasPrefix(peer, "https://") {
			peer = "http://" + peer
		}
		peers = append(peers, peer)
	}
	return peers
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
