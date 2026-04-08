package source

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//go:embed default_sources.json
var defaultSourcesFS embed.FS

// SourceManager 媒体源管理器
type SourceManager struct {
	sources    []MediaSource
	subscriptions []Subscription
	cachePath  string
	mu         sync.RWMutex
}

// Subscription 订阅配置
type Subscription struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updatedAt"`
}

// SourceConfig 媒体源配置文件
type SourceConfig struct {
	Subscriptions []Subscription `json:"subscriptions"`
	Sources       []MediaSource  `json:"sources"`
}

// NewSourceManager 创建媒体源管理器
func NewSourceManager() *SourceManager {
	home, _ := os.UserHomeDir()
	cachePath := filepath.Join(home, ".goani", "sources_cache.json")

	sm := &SourceManager{
		cachePath: cachePath,
	}
	sm.load()
	return sm
}

// load 加载媒体源（缓存优先，无缓存则用默认）
func (sm *SourceManager) load() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 1. 尝试加载缓存
	if sm.loadCache() {
		return
	}

	// 2. 加载默认源
	sm.sources = sm.loadDefaultSources()
	sm.subscriptions = []Subscription{}
}

// loadCache 加载缓存的配置
func (sm *SourceManager) loadCache() bool {
	data, err := os.ReadFile(sm.cachePath)
	if err != nil {
		return false
	}

	var cfg SourceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}

	sm.sources = cfg.Sources
	sm.subscriptions = cfg.Subscriptions
	return len(sm.sources) > 0
}

// loadDefaultSources 加载默认媒体源
func (sm *SourceManager) loadDefaultSources() []MediaSource {
	data, err := defaultSourcesFS.ReadFile("default_sources.json")
	if err != nil {
		return nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	return cfg.ExportedMediaSourceDataList.MediaSources
}

// saveCache 保存缓存
func (sm *SourceManager) saveCache() error {
	configDir := filepath.Dir(sm.cachePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	cfg := SourceConfig{
		Subscriptions: sm.subscriptions,
		Sources:       sm.sources,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.cachePath, data, 0644)
}

// GetAll 获取所有媒体源
func (sm *SourceManager) GetAll() []MediaSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sources
}

// GetByIndex 获取指定索引的媒体源
func (sm *SourceManager) GetByIndex(index int) *MediaSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if index < 0 || index >= len(sm.sources) {
		return nil
	}
	return &sm.sources[index]
}

// GetByName 根据名称获取媒体源
func (sm *SourceManager) GetByName(name string) *MediaSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for i := range sm.sources {
		if sm.sources[i].Arguments.Name == name {
			return &sm.sources[i]
		}
	}
	return nil
}

// Count 获取媒体源数量
func (sm *SourceManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sources)
}

// GetSubscriptions 获取订阅列表
func (sm *SourceManager) GetSubscriptions() []Subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.subscriptions
}

// Subscribe 添加订阅
func (sm *SourceManager) Subscribe(url, name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查是否已订阅
	for _, sub := range sm.subscriptions {
		if sub.URL == url {
			return fmt.Errorf("已订阅该源")
		}
	}

	// 获取远程配置
	sources, err := sm.fetchFromURL(url)
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}

	// 添加订阅
	sub := Subscription{
		URL:       url,
		Name:      name,
		UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	sm.subscriptions = append(sm.subscriptions, sub)
	sm.sources = append(sm.sources, sources...)

	return sm.saveCache()
}

// Unsubscribe 取消订阅
func (sm *SourceManager) Unsubscribe(url string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 移除订阅
	for i, sub := range sm.subscriptions {
		if sub.URL == url {
			sm.subscriptions = append(sm.subscriptions[:i], sm.subscriptions[i+1:]...)
			break
		}
	}

	// 重新加载：默认源 + 剩余订阅
	sm.sources = sm.loadDefaultSources()
	for _, sub := range sm.subscriptions {
		sources, _ := sm.fetchFromURL(sub.URL)
		sm.sources = append(sm.sources, sources...)
	}

	return sm.saveCache()
}

// Refresh 刷新订阅
func (sm *SourceManager) Refresh() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.subscriptions) == 0 {
		return nil
	}

	// 重新获取所有订阅
	sm.sources = sm.loadDefaultSources()
	for i := range sm.subscriptions {
		sources, err := sm.fetchFromURL(sm.subscriptions[i].URL)
		if err != nil {
			continue
		}
		sm.subscriptions[i].UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
		sm.sources = append(sm.sources, sources...)
	}

	return sm.saveCache()
}

// Reset 重置为默认源
func (sm *SourceManager) Reset() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sources = sm.loadDefaultSources()
	sm.subscriptions = []Subscription{}

	// 删除缓存文件
	os.Remove(sm.cachePath)

	return nil
}

// fetchFromURL 从URL获取媒体源配置
func (sm *SourceManager) fetchFromURL(url string) ([]MediaSource, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg.ExportedMediaSourceDataList.MediaSources, nil
}

// GetCachePath 获取缓存文件路径
func GetCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goani", "sources_cache.json")
}
