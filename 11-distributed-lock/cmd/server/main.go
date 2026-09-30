package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"distributed-lock/lock"
)

type server struct {
	mutex *lock.Lock
}

func main() {
	address := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	s := &server{mutex: lock.New()}

	http.HandleFunc("/acquire", s.acquire)
	http.HandleFunc("/renew", s.renew)
	http.HandleFunc("/release", s.release)

	log.Printf("distributed lock listening on %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func (s *server) acquire(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")
	ttl, err := ttlFromRequest(r)
	if err != nil || owner == "" {
		http.Error(w, "owner and valid ttl_ms are required", http.StatusBadRequest)
		return
	}

	token, acquired := s.mutex.Acquire(owner, ttl)
	writeJSON(w, map[string]any{"acquired": acquired, "token": token})
}

func (s *server) renew(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")
	token, tokenErr := strconv.ParseUint(r.URL.Query().Get("token"), 10, 64)
	ttl, ttlErr := ttlFromRequest(r)
	if owner == "" || tokenErr != nil || ttlErr != nil {
		http.Error(w, "owner, token and valid ttl_ms are required", http.StatusBadRequest)
		return
	}

	renewed := s.mutex.Renew(owner, token, ttl)
	writeJSON(w, map[string]any{"renewed": renewed})
}

func (s *server) release(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")
	token, err := strconv.ParseUint(r.URL.Query().Get("token"), 10, 64)
	if owner == "" || err != nil {
		http.Error(w, "owner and valid token are required", http.StatusBadRequest)
		return
	}

	released := s.mutex.Release(owner, token)
	writeJSON(w, map[string]any{"released": released})
}

func ttlFromRequest(r *http.Request) (time.Duration, error) {
	milliseconds, err := strconv.ParseInt(r.URL.Query().Get("ttl_ms"), 10, 64)
	if err != nil || milliseconds <= 0 {
		if err == nil {
			err = errors.New("ttl_ms must be positive")
		}
		return 0, err
	}
	return time.Duration(milliseconds) * time.Millisecond, nil
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
