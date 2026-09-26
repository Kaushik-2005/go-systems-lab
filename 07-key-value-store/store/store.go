package store

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type record struct {
	Operation string `json:"operation"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
}

type Store struct {
	mu     sync.RWMutex
	values map[string]string
	file   *os.File
	path   string
}

func New() *Store {
	return &Store{
		values: make(map[string]string),
	}
}

func Open(path string) (*Store, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}

	database := &Store{
		values: make(map[string]string),
		file:   file,
		path:   path,
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry record
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			file.Close()
			return nil, err
		}

		database.applyRecord(entry)
	}

	if err := scanner.Err(); err != nil {
		file.Close()
		return nil, err
	}

	return database, nil
}

func (s *Store) Compact() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file == nil || s.path == "" {
		return nil
	}

	tempFile, err := os.CreateTemp(
		filepath.Dir(s.path),
		filepath.Base(s.path)+".compact-*",
	)
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()

	cleanup := func() {
		tempFile.Close()
		os.Remove(tempPath)
	}

	keys := make([]string, 0, len(s.values))
	for key := range s.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	writer := bufio.NewWriter(tempFile)

	for _, key := range keys {
		entry := record{
			Operation: "put",
			Key:       key,
			Value:     s.values[key],
		}

		data, err := json.Marshal(entry)
		if err != nil {
			cleanup()
			return err
		}

		if _, err := writer.Write(append(data, '\n')); err != nil {
			cleanup()
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		cleanup()
		return err
	}

	if err := tempFile.Sync(); err != nil {
		cleanup()
		return err
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}

	if err := s.file.Close(); err != nil {
		os.Remove(tempPath)
		return err
	}

	if err := os.Rename(tempPath, s.path); err != nil {
		os.Remove(tempPath)
		return err
	}

	s.file, err = os.OpenFile(
		s.path,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Put(key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.appendRecord(record{
		Operation: "put",
		Key:       key,
		Value:     value,
	})
	if err != nil {
		return err
	}

	s.values[key] = value
	return nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, found := s.values[key]
	return value, found
}

func (s *Store) Delete(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.values[key]; !found {
		return false, nil
	}

	err := s.appendRecord(record{
		Operation: "delete",
		Key:       key,
	})
	if err != nil {
		return false, err
	}

	delete(s.values, key)
	return true, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file == nil {
		return nil
	}

	err := s.file.Sync()
	if closeErr := s.file.Close(); err == nil {
		err = closeErr
	}

	s.file = nil
	return err
}

func (s *Store) appendRecord(entry record) error {
	if s.file == nil {
		return nil
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	if _, err := s.file.Write(data); err != nil {
		return err
	}

	return s.file.Sync()
}

func (s *Store) applyRecord(entry record) {
	switch entry.Operation {
	case "put":
		s.values[entry.Key] = entry.Value

	case "delete":
		delete(s.values, entry.Key)
	}
}
