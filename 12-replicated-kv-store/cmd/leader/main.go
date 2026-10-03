package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"replicated-kv-store/store"
)

type server struct {
	store *store.Store
}

func main() {
	address := flag.String("addr", ":9100", "HTTP listen address")
	flag.Parse()

	s := &server{store: store.New()}
	http.HandleFunc("/put", s.put)
	http.HandleFunc("/delete", s.delete)
	http.HandleFunc("/get", s.get)
	http.HandleFunc("/records", s.records)
	http.HandleFunc("/status", s.status)

	log.Printf("leader listening on %s", *address)
	log.Fatal(http.ListenAndServe(*address, nil))
}

func (s *server) put(w http.ResponseWriter, r *http.Request) {
	key, value := r.URL.Query().Get("key"), r.URL.Query().Get("value")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	writeJSON(w, s.store.Put(key, value))
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	writeJSON(w, s.store.Delete(key))
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	value, found := s.store.Get(r.URL.Query().Get("key"))
	writeJSON(w, map[string]any{"value": value, "found": found, "sequence": s.store.LastSequence()})
}

func (s *server) records(w http.ResponseWriter, r *http.Request) {
	sequence, err := strconv.ParseUint(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		http.Error(w, "valid from sequence is required", http.StatusBadRequest)
		return
	}

	writeJSON(w, s.store.RecordsAfter(sequence))
}

func (s *server) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]uint64{"sequence": s.store.LastSequence()})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
