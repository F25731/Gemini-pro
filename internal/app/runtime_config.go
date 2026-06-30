package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type RuntimeConfig struct {
	BananaAPIKey string `json:"bananaApiKey"`
}

type RuntimeStore struct {
	path string
	mu   sync.RWMutex
	cfg  RuntimeConfig
}

func NewRuntimeStore(path string, fallback RuntimeConfig) *RuntimeStore {
	store := &RuntimeStore{path: path, cfg: fallback}
	store.Load()
	return store
}

func (s *RuntimeStore) Load() {
	body, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var cfg RuntimeConfig
	if json.Unmarshal(body, &cfg) != nil {
		return
	}
	s.mu.Lock()
	if strings.TrimSpace(cfg.BananaAPIKey) != "" {
		s.cfg.BananaAPIKey = strings.TrimSpace(cfg.BananaAPIKey)
	}
	s.mu.Unlock()
}

func (s *RuntimeStore) Get() RuntimeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *RuntimeStore) Save(cfg RuntimeConfig) error {
	cfg.BananaAPIKey = strings.TrimSpace(cfg.BananaAPIKey)
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	body, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(s.path, body, 0600)
}
