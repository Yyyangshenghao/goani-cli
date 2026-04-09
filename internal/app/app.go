package app

import (
	"fmt"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/settings"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

// App 应用核心，管理全局资源和配置
type App struct {
	Settings      *settings.Config
	PlayerConfig  *player.Config
	SourceManager *source.SourceManager
}

// New 创建应用实例
func New() *App {
	cfg, err := settings.Load()
	if err != nil {
		cfg = settings.DefaultConfig()
	}

	return &App{
		Settings:      cfg,
		PlayerConfig:  playerConfigFromSettings(cfg),
		SourceManager: source.NewSourceManager(cfg),
	}
}

// GetSource 获取指定索引的媒体源
func (a *App) GetSource(index int) *webselector.WebSelectorSource {
	ms := a.SourceManager.GetByIndex(index)
	if ms == nil {
		return nil
	}
	return webselector.New(*ms)
}

// GetFirstSource 获取第一个媒体源
func (a *App) GetFirstSource() *webselector.WebSelectorSource {
	return a.GetSource(0)
}

// GetSourceByName 根据名称获取媒体源
func (a *App) GetSourceByName(name string) *webselector.WebSelectorSource {
	ms := a.SourceManager.GetByName(name)
	if ms == nil {
		return nil
	}
	return webselector.New(*ms)
}

// GetPlayer 获取播放时应使用的播放器。
// 如果当前没有默认播放器，会优先从已配置路径里挑一个可用播放器并写回配置。
func (a *App) GetPlayer() (player.Player, error) {
	if strings.TrimSpace(a.PlayerConfig.DefaultPlayer) == "" {
		if configured := player.FirstConfiguredAvailablePlayer(a.PlayerConfig.Paths); configured != nil {
			a.PlayerConfig.SetDefaultPlayer(configured.Name())
			if err := a.SaveConfig(); err != nil {
				return nil, fmt.Errorf("已找到可用播放器 %s，但保存默认播放器失败: %w", configured.Name(), err)
			}
		}
	}

	pm := player.NewManagerWithConfig(a.PlayerConfig.DefaultPlayer, a.PlayerConfig.Paths)
	return pm.GetFirst(), nil
}

// SaveConfig 保存配置
func (a *App) SaveConfig() error {
	a.Settings.Player.Default = a.PlayerConfig.DefaultPlayer
	a.Settings.Player.Paths = map[string]string{}
	for name, path := range a.PlayerConfig.Paths {
		a.Settings.Player.Paths[name] = path
	}
	return a.Settings.Save()
}

// GetAllSources 获取所有媒体源
func (a *App) GetAllSources() []source.MediaSource {
	return a.SourceManager.GetAll()
}

func playerConfigFromSettings(cfg *settings.Config) *player.Config {
	pc := player.DefaultConfig()
	pc.DefaultPlayer = cfg.Player.Default
	for name, path := range cfg.Player.Paths {
		pc.Paths[name] = path
	}
	return pc
}
