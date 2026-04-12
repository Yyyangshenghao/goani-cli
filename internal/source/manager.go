package source

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/settings"
)

// SourceManager 媒体源管理器
type SourceManager struct {
	allSources       []MediaSource
	enabledSources   []MediaSource
	allSourcesByName map[string]MediaSource
	priorityByName   map[string]int
	subscriptions    []Subscription
	config           *settings.Config
	cache            *sourceCacheStore
	preferenceStore  *sourcePreferenceStore
	preferences      SourcePreferenceFile
	fetcher          *sourceFetcher
	mu               sync.RWMutex
}

// NewSourceManager 创建媒体源管理器
func NewSourceManager(cfg *settings.Config) *SourceManager {
	cachePath, _ := settings.GetSourceCachePath()
	preferencePath, _ := settings.GetSourcePreferencesPath()

	sm := &SourceManager{
		config:          cfg,
		cache:           newSourceCacheStore(cachePath),
		preferenceStore: newSourcePreferenceStore(preferencePath),
		fetcher:         newSourceFetcher(30 * time.Second),
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

	sm.loadPreferencesLocked()

	// 1. 尝试加载缓存
	if sm.loadCache() {
		sm.refreshPreferencesLocked()
		return
	}

	// 2. 刷新当前订阅
	subscriptions, sources, ok := sm.fetchSubscriptionsLocked(sm.subscriptions)
	if !ok {
		sm.allSources = []MediaSource{}
		sm.refreshPreferencesLocked()
		return
	}
	sm.subscriptions = subscriptions
	sm.allSources = cloneMediaSources(sources)
	sm.refreshPreferencesLocked()
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

	sm.allSources = cloneMediaSources(cfg.Sources)
	if len(cfg.Subscriptions) > 0 {
		sm.subscriptions = cloneSubscriptions(cfg.Subscriptions)
		sm.config.Sources.Subscriptions = subscriptionsToSettings(cfg.Subscriptions)
		_ = sm.config.Save()
	}
	return len(sm.allSources) > 0
}

// saveCache 保存缓存
func (sm *SourceManager) saveCache() error {
	return sm.cache.Save(SourceCache{
		Subscriptions: cloneSubscriptions(sm.subscriptions),
		Sources:       cloneMediaSources(sm.allSources),
	})
}

// GetAll 获取所有媒体源
func (sm *SourceManager) GetAll() []MediaSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return cloneMediaSources(sm.allSources)
}

// GetEnabled 获取所有启用中的媒体源
func (sm *SourceManager) GetEnabled() []MediaSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return cloneMediaSources(sm.enabledSources)
}

// GetByIndex 获取指定索引的媒体源
func (sm *SourceManager) GetByIndex(index int) *MediaSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if index < 0 || index >= len(sm.enabledSources) {
		return nil
	}
	source := sm.enabledSources[index]
	return &source
}

// GetByName 根据名称获取媒体源
func (sm *SourceManager) GetByName(name string) *MediaSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	source, ok := sm.allSourcesByName[name]
	if !ok {
		return nil
	}
	return &source
}

// Count 获取媒体源数量
func (sm *SourceManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.allSources)
}

// EnabledCount 获取启用中的媒体源数量
func (sm *SourceManager) EnabledCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.enabledSources)
}

// GetChannelPriorityByName 获取渠道优先级。
func (sm *SourceManager) GetChannelPriorityByName(name string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.priorityByName[name]
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
	sm.allSources = append(sm.allSources, cloneMediaSources(sources)...)
	sm.refreshPreferencesLocked()

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
	sm.allSources = cloneMediaSources(sources)
	sm.refreshPreferencesLocked()
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
		sm.allSources = []MediaSource{}
		sm.refreshPreferencesLocked()
		return sm.saveLocked()
	}

	refreshedSubscriptions, sources, ok := sm.fetchSubscriptionsLocked(nextSubscriptions)
	if !ok {
		return fmt.Errorf("无法刷新剩余订阅，已保留原有配置")
	}

	sm.subscriptions = refreshedSubscriptions
	sm.allSources = cloneMediaSources(sources)
	sm.refreshPreferencesLocked()
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
	sm.allSources = cloneMediaSources(sources)
	sm.refreshPreferencesLocked()
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
	sm.allSources = cloneMediaSources(sources)
	sm.refreshPreferencesLocked()

	// 删除旧缓存文件后保存新状态
	_ = sm.cache.Remove()
	return sm.saveLocked()
}

// GetCachePath 获取缓存文件路径
func GetCachePath() string {
	path, _ := settings.GetSourceCachePath()
	return path
}

// GetPreferencesPath 获取片源渠道偏好文件路径
func GetPreferencesPath() string {
	path, _ := settings.GetSourcePreferencesPath()
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
	if err := sm.saveCache(); err != nil {
		return err
	}
	return sm.savePreferencesLocked()
}

// ListChannels 返回所有已加载渠道及其启用和探测状态。
func (sm *SourceManager) ListChannels() []SourceChannelState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	prefByID := make(map[string]SourcePreference, len(sm.preferences.Sources))
	for _, pref := range sm.preferences.Sources {
		prefByID[pref.ID] = pref
	}

	items := make([]SourceChannelState, 0, len(sm.allSources))
	for _, ms := range sm.allSources {
		id := sourcePreferenceID(ms)
		pref, ok := prefByID[id]
		enabled := true
		if ok {
			enabled = pref.Enabled
		}

		items = append(items, SourceChannelState{
			ID:          id,
			Name:        displaySourcePreferenceName(ms),
			Description: ms.Arguments.Description,
			Enabled:     enabled,
			Priority:    pref.Priority,
			LastDoctor:  cloneDoctorSnapshot(pref.LastDoctor),
		})
	}

	sortSourceChannelStates(items)
	return items
}

// SetChannelEnabled 设置单个渠道开关状态。
func (sm *SourceManager) SetChannelEnabled(id string, enabled bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	index := sm.preferenceIndexLocked(id)
	if index < 0 {
		return fmt.Errorf("未找到片源渠道: %s", id)
	}

	sm.preferences.Sources[index].Enabled = enabled
	sm.applyEnabledSourcesLocked()
	return sm.savePreferencesLocked()
}

// SetAllChannelsEnabled 批量设置所有渠道的开关状态。
func (sm *SourceManager) SetAllChannelsEnabled(enabled bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i := range sm.preferences.Sources {
		sm.preferences.Sources[i].Enabled = enabled
	}
	sm.applyEnabledSourcesLocked()
	return sm.savePreferencesLocked()
}

// EnableOnlyChannels 仅保留指定渠道为启用状态。
func (sm *SourceManager) EnableOnlyChannels(ids []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	allowed := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		allowed[id] = struct{}{}
	}

	for i := range sm.preferences.Sources {
		_, ok := allowed[sm.preferences.Sources[i].ID]
		sm.preferences.Sources[i].Enabled = ok
	}
	sm.applyEnabledSourcesLocked()
	return sm.savePreferencesLocked()
}

// EnableLastWorkingChannels 仅启用最近一次探测成功拿到视频直链的渠道。
func (sm *SourceManager) EnableLastWorkingChannels() (int, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	enabledCount := 0
	for i := range sm.preferences.Sources {
		working := sm.preferences.Sources[i].LastDoctor != nil && sm.preferences.Sources[i].LastDoctor.SuccessfulRuns > 0
		sm.preferences.Sources[i].Enabled = working
		if working {
			enabledCount++
		}
	}
	if enabledCount == 0 {
		return 0, fmt.Errorf("最近一次探测里没有可直链播放的源，未修改启用状态")
	}

	sm.applyEnabledSourcesLocked()
	if err := sm.savePreferencesLocked(); err != nil {
		return 0, err
	}
	return enabledCount, nil
}

// UpdateDoctorSnapshots 写回渠道诊断结果。
func (sm *SourceManager) UpdateDoctorSnapshots(updates []ChannelDoctorUpdate) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, update := range updates {
		index := sm.preferenceIndexLocked(update.ID)
		if index < 0 {
			sm.preferences.Sources = append(sm.preferences.Sources, SourcePreference{
				ID:      update.ID,
				Name:    strings.TrimSpace(update.Name),
				Enabled: true,
			})
			index = len(sm.preferences.Sources) - 1
		}
		if strings.TrimSpace(update.Name) != "" {
			sm.preferences.Sources[index].Name = strings.TrimSpace(update.Name)
		}
		sm.preferences.Sources[index].Priority = update.Priority
		sm.preferences.Sources[index].Enabled = update.Enabled
		snapshot := update.Snapshot
		sm.preferences.Sources[index].LastDoctor = cloneDoctorSnapshot(&snapshot)
	}

	sm.applyEnabledSourcesLocked()
	return sm.savePreferencesLocked()
}

func (sm *SourceManager) loadPreferencesLocked() {
	file, err := sm.preferenceStore.Load()
	if err != nil {
		sm.preferences = SourcePreferenceFile{}
		return
	}
	sm.preferences = SourcePreferenceFile{
		Sources: make([]SourcePreference, 0, len(file.Sources)),
	}
	for _, pref := range file.Sources {
		sm.preferences.Sources = append(sm.preferences.Sources, cloneSourcePreference(pref))
	}
}

func (sm *SourceManager) refreshPreferencesLocked() {
	sm.preferences = syncSourcePreferences(sm.allSources, sm.preferences)
	sm.applyEnabledSourcesLocked()
	_ = sm.savePreferencesLocked()
}

func (sm *SourceManager) applyEnabledSourcesLocked() {
	prefByID := make(map[string]SourcePreference, len(sm.preferences.Sources))
	for _, pref := range sm.preferences.Sources {
		prefByID[pref.ID] = pref
	}

	sm.allSourcesByName = make(map[string]MediaSource, len(sm.allSources))
	sm.priorityByName = make(map[string]int, len(sm.allSources))
	sm.enabledSources = make([]MediaSource, 0, len(sm.allSources))
	for _, ms := range sm.allSources {
		name := strings.TrimSpace(ms.Arguments.Name)
		if name != "" {
			sm.allSourcesByName[name] = ms
			if pref, ok := prefByID[sourcePreferenceID(ms)]; ok {
				sm.priorityByName[name] = pref.Priority
			}
		}

		pref, ok := prefByID[sourcePreferenceID(ms)]
		if ok && !pref.Enabled {
			continue
		}
		sm.enabledSources = append(sm.enabledSources, ms)
	}
	sortMediaSourcesByPreference(sm.enabledSources, prefByID)
}

func (sm *SourceManager) savePreferencesLocked() error {
	return sm.preferenceStore.Save(sm.preferences)
}

func (sm *SourceManager) preferenceIndexLocked(id string) int {
	for i := range sm.preferences.Sources {
		if sm.preferences.Sources[i].ID == id {
			return i
		}
	}
	return -1
}

func (sm *SourceManager) preferenceByIDLocked(id string) (SourcePreference, bool) {
	for i := range sm.preferences.Sources {
		if sm.preferences.Sources[i].ID == id {
			return sm.preferences.Sources[i], true
		}
	}
	return SourcePreference{}, false
}

func cloneMediaSources(items []MediaSource) []MediaSource {
	if len(items) == 0 {
		return nil
	}

	cloned := make([]MediaSource, len(items))
	copy(cloned, items)
	return cloned
}

func sortMediaSourcesByPreference(items []MediaSource, prefByID map[string]SourcePreference) {
	sort.SliceStable(items, func(i, j int) bool {
		leftID := sourcePreferenceID(items[i])
		rightID := sourcePreferenceID(items[j])
		leftPriority := prefByID[leftID].Priority
		rightPriority := prefByID[rightID].Priority
		if leftPriority != rightPriority {
			return leftPriority > rightPriority
		}
		return strings.TrimSpace(items[i].Arguments.Name) < strings.TrimSpace(items[j].Arguments.Name)
	})
}

func sortSourceChannelStates(items []SourceChannelState) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority > items[j].Priority
		}
		return items[i].Name < items[j].Name
	})
}
