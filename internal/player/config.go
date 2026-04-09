package player

import "github.com/Yyyangshenghao/goani-cli/internal/settings"

// Config 播放器配置
type Config struct {
	DefaultPlayer string            `json:"defaultPlayer"`
	Paths         map[string]string `json:"paths"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		DefaultPlayer: "",
		Paths:         map[string]string{},
	}
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	cfg, err := settings.Load()
	if err != nil {
		return DefaultConfig(), nil
	}

	return fromSettings(cfg.Player), nil
}

// Save 保存配置
func (c *Config) Save() error {
	appCfg, err := settings.Load()
	if err != nil {
		return err
	}

	appCfg.Player = toSettings(c)
	return appCfg.Save()
}

// SetPlayer 设置播放器
func (c *Config) SetPlayer(name, path string) {
	if c.Paths == nil {
		c.Paths = map[string]string{}
	}
	c.Paths[name] = path
	c.DefaultPlayer = name
}

// SetPlayerPath 仅更新播放器路径，不会隐式修改默认播放器。
func (c *Config) SetPlayerPath(name, path string) {
	if c.Paths == nil {
		c.Paths = map[string]string{}
	}
	c.Paths[name] = path
}

// SetDefaultPlayer 设置默认播放器
func (c *Config) SetDefaultPlayer(name string) {
	c.DefaultPlayer = name
}

// GetPath 获取播放器路径
func (c *Config) GetPath(name string) string {
	if c.Paths == nil {
		return ""
	}
	return c.Paths[name]
}

func fromSettings(cfg settings.PlayerConfig) *Config {
	paths := map[string]string{}
	for name, path := range cfg.Paths {
		paths[name] = path
	}

	return &Config{
		DefaultPlayer: cfg.Default,
		Paths:         paths,
	}
}

func toSettings(cfg *Config) settings.PlayerConfig {
	paths := map[string]string{}
	if cfg != nil {
		for name, path := range cfg.Paths {
			paths[name] = path
		}
	}

	defaultPlayer := ""
	if cfg != nil {
		defaultPlayer = cfg.DefaultPlayer
	}

	return settings.PlayerConfig{
		Default: defaultPlayer,
		Paths:   paths,
	}
}
