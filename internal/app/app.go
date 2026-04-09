package app

import (
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

// GetPlayer 获取配置的播放器
func (a *App) GetPlayer() player.Player {
	pm := player.NewManagerWithConfig(a.PlayerConfig.DefaultPlayer, a.PlayerConfig.Paths)
	return pm.GetFirst()
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
