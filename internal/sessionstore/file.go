package sessionstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type entry struct {
	Data      []byte `json:"d"`
	ExpiresAt int64  `json:"e"`
}

type FileStorage struct {
	dir string
	mu  sync.RWMutex
}

func New(dir string) *FileStorage {
	os.MkdirAll(dir, 0755)
	return &FileStorage{dir: dir}
}

func (s *FileStorage) path(key string) string {
	return filepath.Join(s.dir, key+".json")
}

func (s *FileStorage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path(key))
	if err != nil {
		return nil, nil
	}

	var e entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, nil
	}

	if e.ExpiresAt > 0 && time.Now().Unix() > e.ExpiresAt {
		os.Remove(s.path(key))
		return nil, nil
	}

	return e.Data, nil
}

func (s *FileStorage) Set(key string, val []byte, exp time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := entry{Data: val}
	if exp > 0 {
		e.ExpiresAt = time.Now().Add(exp).Unix()
	}

	data, err := json.Marshal(&e)
	if err != nil {
		return err
	}

	return os.WriteFile(s.path(key), data, 0644)
}

func (s *FileStorage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.path(key))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *FileStorage) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		os.Remove(filepath.Join(s.dir, e.Name()))
	}
	return nil
}

func (s *FileStorage) Close() error {
	return nil
}
