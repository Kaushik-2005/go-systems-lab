package store

import (
	"errors"
	"sync"
)

var ErrOutOfOrder = errors.New("replication record is out of order")

type Record struct {
	Sequence uint64 `json:"sequence"`
	Key      string `json:"key"`
	Value    string `json:"value,omitempty"`
	Deleted  bool   `json:"deleted"`
}

type Store struct {
	mu      sync.RWMutex
	values  map[string]string
	lastSeq uint64
	log     []Record
}

func New() *Store {
	return &Store{values: make(map[string]string)}
}

func (s *Store) Put(key, value string) Record {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := Record{Sequence: s.lastSeq + 1, Key: key, Value: value}
	s.applyLocked(record)
	s.log = append(s.log, record)
	return record
}

func (s *Store) Delete(key string) Record {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := Record{Sequence: s.lastSeq + 1, Key: key, Deleted: true}
	s.applyLocked(record)
	s.log = append(s.log, record)
	return record
}

func (s *Store) Apply(record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.Sequence <= s.lastSeq {
		return nil
	}
	if record.Sequence != s.lastSeq+1 {
		return ErrOutOfOrder
	}

	s.applyLocked(record)
	return nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, found := s.values[key]
	return value, found
}

func (s *Store) RecordsAfter(sequence uint64) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sequence >= s.lastSeq {
		return nil
	}

	start := int(sequence)
	result := make([]Record, len(s.log)-start)
	copy(result, s.log[start:])
	return result
}

func (s *Store) LastSequence() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSeq
}

func (s *Store) applyLocked(record Record) {
	s.lastSeq = record.Sequence
	if record.Deleted {
		delete(s.values, record.Key)
		return
	}
	s.values[record.Key] = record.Value
}
