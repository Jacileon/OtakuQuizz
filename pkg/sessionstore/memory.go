package sessionstore

import (
	"sync"
	"time"
)

type MemoryStorage struct {
	mu      sync.RWMutex
	entries map[string]memoryEntry
}

type memoryEntry struct {
	Data      []byte
	ExpiresAt time.Time
}

func NewMemory() *MemoryStorage {
	ms := &MemoryStorage{entries: make(map[string]memoryEntry)}
	go ms.cleanup()
	return ms
}

func (s *MemoryStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.entries {
			if !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *MemoryStorage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.entries[key]
	if !ok {
		return nil, nil
	}
	if !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt) {
		return nil, nil
	}
	return e.Data, nil
}

func (s *MemoryStorage) Set(key string, val []byte, exp time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := memoryEntry{Data: val}
	if exp > 0 {
		e.ExpiresAt = time.Now().Add(exp)
	}
	s.entries[key] = e
	return nil
}

func (s *MemoryStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
	return nil
}

func (s *MemoryStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make(map[string]memoryEntry)
	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}
