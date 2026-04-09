package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultSourceURL 默认媒体源订阅地址
const DefaultSourceURL = "https://sub.creamycake.org/v1/css1.json"

// Config 应用统一配置
type Config struct {
	Player  PlayerConfig `json:"player"`
	Sources SourceConfig `json:"sources"`
}

// PlayerConfig 播放器配置
type PlayerConfig struct {
	Default string            `json:"default"`
	Paths   map[string]string `json:"paths,omitempty"`
}

// SourceConfig 片源配置
type SourceConfig struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

// Subscription 订阅配置
type Subscription struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Player: PlayerConfig{
			Paths: map[string]string{},
		},
		Sources: SourceConfig{
			Subscriptions: DefaultSubscriptions(),
		},
	}
}

// DefaultSubscriptions 默认订阅列表
func DefaultSubscriptions() []Subscription {
	return []Subscription{
		{
			URL:  DefaultSourceURL,
			Name: "默认源",
		},
	}
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// GetSourceCachePath 获取片源缓存路径
func GetSourceCachePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sources_cache.json"), nil
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".goani")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configDir, nil
}

// Load 加载统一配置
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Player.Paths == nil {
		cfg.Player.Paths = map[string]string{}
	}
	if len(cfg.Sources.Subscriptions) == 0 {
		cfg.Sources.Subscriptions = DefaultSubscriptions()
	}

	return cfg, nil
}

// Save 保存统一配置
func (c *Config) Save() error {
	if c.Player.Paths == nil {
		c.Player.Paths = map[string]string{}
	}
	if len(c.Sources.Subscriptions) == 0 {
		c.Sources.Subscriptions = DefaultSubscriptions()
	}

	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// CloneSubscriptions 复制订阅列表
func CloneSubscriptions(subs []Subscription) []Subscription {
	return cloneSubscriptions(subs)
}

func cloneSubscriptions(subs []Subscription) []Subscription {
	if len(subs) == 0 {
		return nil
	}

	cloned := make([]Subscription, len(subs))
	copy(cloned, subs)
	return cloned
}
