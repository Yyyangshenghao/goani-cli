package source

import (
	"fmt"
	"sync"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/settings"
)

// SourceManager 媒体源管理器
type SourceManager struct {
	sources       []MediaSource
	subscriptions []Subscription
	config        *settings.Config
	cache         *sourceCacheStore
	fetcher       *sourceFetcher
	mu            sync.RWMutex
}

// NewSourceManager 创建媒体源管理器
func NewSourceManager(cfg *settings.Config) *SourceManager {
	cachePath, _ := settings.GetSourceCachePath()

	sm := &SourceManager{
		config:  cfg,
		cache:   newSourceCacheStore(cachePath),
		fetcher: newSourceFetcher(30 * time.Second),
	}
	sm.load()
	return sm
}

// load 加载媒体源（缓存优先，无缓存则订阅默认源）
func (sm *SourceManager) load() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.subscriptions = subscriptionsFromSettings(sm.config.Sources.Subscriptions)
	if len(sm.subscriptions) == 0 {
		sm.subscriptions = defaultSubscriptions()
	}

	// 1. 尝试加载缓存
	if sm.loadCache() {
		return
	}

	// 2. 刷新当前订阅
	subscriptions, sources, ok := sm.fetchSubscriptionsLocked(sm.subscriptions)
	if !ok {
		sm.sources = []MediaSource{}
		return
	}
	sm.subscriptions = subscriptions
	sm.sources = sources
	_ = sm.saveLocked()
}

// loadCache 加载缓存的配置
func (sm *SourceManager) loadCache() bool {
	cfg, err := sm.cache.Load()
	if err != nil {
		return false
	}

	if !sameSubscriptions(cfg.Subscriptions, sm.subscriptions) {
		return false
	}

	sm.sources = cfg.Sources
	if len(cfg.Subscriptions) > 0 {
		sm.subscriptions = cloneSubscriptions(cfg.Subscriptions)
		sm.config.Sources.Subscriptions = subscriptionsToSettings(cfg.Subscriptions)
		_ = sm.config.Save()
	}
	return len(sm.sources) > 0
}

// saveCache 保存缓存
func (sm *SourceManager) saveCache() error {
	return sm.cache.Save(SourceCache{
		Subscriptions: cloneSubscriptions(sm.subscriptions),
		Sources:       sm.sources,
	})
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
	return cloneSubscriptions(sm.subscriptions)
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
	sources, err := sm.fetcher.FetchURL(url)
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

	return sm.saveLocked()
}

// UpsertSubscription 新增或更新订阅，并在保存前重新抓取订阅内容。
// previousURL 为空表示新增；否则会把对应订阅更新为新的名称或 URL。
func (sm *SourceManager) UpsertSubscription(previousURL, url, name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	nextSubscriptions := cloneSubscriptions(sm.subscriptions)
	targetIndex := -1

	if previousURL != "" {
		for i, sub := range nextSubscriptions {
			if sub.URL == previousURL {
				targetIndex = i
				break
			}
		}
		if targetIndex < 0 {
			return fmt.Errorf("未找到要编辑的订阅")
		}
	}

	for i, sub := range nextSubscriptions {
		if sub.URL != url {
			continue
		}
		if previousURL != "" && i == targetIndex {
			continue
		}
		return fmt.Errorf("已订阅该源")
	}

	if name == "" {
		name = "自定义源"
	}

	if previousURL == "" {
		nextSubscriptions = append(nextSubscriptions, Subscription{
			URL:       url,
			Name:      name,
			UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
		})
	} else {
		nextSubscriptions[targetIndex].URL = url
		nextSubscriptions[targetIndex].Name = name
		nextSubscriptions[targetIndex].UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	refreshedSubscriptions, sources, ok := sm.fetchSubscriptionsLocked(nextSubscriptions)
	if !ok {
		return fmt.Errorf("订阅验证失败，已保留原有配置")
	}

	sm.subscriptions = refreshedSubscriptions
	sm.sources = sources
	return sm.saveLocked()
}

// Unsubscribe 取消订阅
func (sm *SourceManager) Unsubscribe(url string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	nextSubscriptions := cloneSubscriptions(sm.subscriptions)

	// 移除订阅
	for i, sub := range nextSubscriptions {
		if sub.URL == url {
			nextSubscriptions = append(nextSubscriptions[:i], nextSubscriptions[i+1:]...)
			break
		}
	}

	if len(nextSubscriptions) == len(sm.subscriptions) {
		return fmt.Errorf("未找到该订阅")
	}

	if len(nextSubscriptions) == 0 {
		sm.subscriptions = nextSubscriptions
		sm.sources = []MediaSource{}
		return sm.saveLocked()
	}

	refreshedSubscriptions, sources, ok := sm.fetchSubscriptionsLocked(nextSubscriptions)
	if !ok {
		return fmt.Errorf("无法刷新剩余订阅，已保留原有配置")
	}

	sm.subscriptions = refreshedSubscriptions
	sm.sources = sources
	return sm.saveLocked()
}

// Refresh 刷新订阅
func (sm *SourceManager) Refresh() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.subscriptions) == 0 {
		return nil
	}

	refreshedSubscriptions, sources, ok := sm.fetchSubscriptionsLocked(sm.subscriptions)
	if !ok {
		return fmt.Errorf("所有订阅刷新失败，已保留原有缓存")
	}

	sm.subscriptions = refreshedSubscriptions
	sm.sources = sources
	return sm.saveLocked()
}

// Reset 重置为默认源
func (sm *SourceManager) Reset() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscriptions := defaultSubscriptions()
	refreshedSubscriptions, sources, ok := sm.fetchSubscriptionsLocked(subscriptions)
	if !ok {
		return fmt.Errorf("默认源刷新失败，已保留原有配置")
	}

	sm.subscriptions = refreshedSubscriptions
	sm.sources = sources

	// 删除旧缓存文件后保存新状态
	_ = sm.cache.Remove()
	return sm.saveLocked()
}

// GetCachePath 获取缓存文件路径
func GetCachePath() string {
	path, _ := settings.GetSourceCachePath()
	return path
}

func (sm *SourceManager) fetchSubscriptionsLocked(subscriptions []Subscription) ([]Subscription, []MediaSource, bool) {
	return sm.fetcher.FetchSubscriptions(subscriptions)
}

func (sm *SourceManager) saveLocked() error {
	sm.config.Sources.Subscriptions = subscriptionsToSettings(sm.subscriptions)
	if err := sm.config.Save(); err != nil {
		return err
	}
	return sm.saveCache()
}
