package source

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SourceCache 片源缓存文件
type SourceCache struct {
	Subscriptions []Subscription `json:"subscriptions"`
	Sources       []MediaSource  `json:"sources"`
}

type sourceCacheStore struct {
	path string
}

func newSourceCacheStore(path string) *sourceCacheStore {
	return &sourceCacheStore{path: path}
}

func (s *sourceCacheStore) Load() (*SourceCache, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	var cache SourceCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func (s *sourceCacheStore) Save(cache SourceCache) error {
	configDir := filepath.Dir(s.path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *sourceCacheStore) Remove() error {
	return os.Remove(s.path)
}
