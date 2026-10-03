package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"replicated-kv-store/store"
)

type server struct {
	store     *store.Store
	leaderURL string
	client    *http.Client
}

func main() {
	address := flag.String("addr", ":9101", "HTTP listen address")
	leaderURL := flag.String("leader", "http://localhost:9100", "leader URL")
	interval := flag.Duration("interval", 500*time.Millisecond, "replication polling interval")
	flag.Parse()

	s := &server{store: store.New(), leaderURL: *leaderURL, client: &http.Client{Timeout: 2 * time.Second}}
	http.HandleFunc("/get", s.get)
	http.HandleFunc("/status", s.status)

	go s.replicateForever(context.Background(), *interval)
	log.Printf("follower listening on %s, pulling from %s", *address, *leaderURL)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func (s *server) replicateForever(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := s.syncOnce(); err != nil {
			log.Printf("replication attempt failed: %v", err)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (s *server) syncOnce() error {
	endpoint := fmt.Sprintf("%s/records?from=%d", s.leaderURL, s.store.LastSequence())
	response, err := s.client.Get(endpoint)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("leader returned %s", response.Status)
	}

	var records []store.Record
	if err := json.NewDecoder(response.Body).Decode(&records); err != nil {
		return err
	}
	for _, record := range records {
		if err := s.store.Apply(record); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	value, found := s.store.Get(r.URL.Query().Get("key"))
	writeJSON(w, map[string]any{"value": value, "found": found, "sequence": s.store.LastSequence()})
}

func (s *server) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]uint64{"sequence": s.store.LastSequence()})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
