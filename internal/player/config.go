package player

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 播放器配置
type Config struct {
	PlayerName string `json:"playerName"`
	PlayerPath string `json:"playerPath"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		PlayerName: "",
		PlayerPath: "",
	}
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".goani")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
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

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save 保存配置
func (c *Config) Save() error {
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

// SetPlayer 设置播放器
func (c *Config) SetPlayer(name, path string) {
	c.PlayerName = name
	c.PlayerPath = path
}
