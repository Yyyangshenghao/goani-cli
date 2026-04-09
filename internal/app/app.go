package app

import (
	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

// App 应用核心，管理全局资源和配置
type App struct {
	PlayerConfig  *player.Config
	SourceManager *source.SourceManager
}

// New 创建应用实例
func New() *App {
	cfg, err := player.LoadConfig()
	if err != nil {
		cfg = player.DefaultConfig()
	}

	return &App{
		PlayerConfig:  cfg,
		SourceManager: source.NewSourceManager(),
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
	pm := player.NewManagerWithConfig(a.PlayerConfig.PlayerName, a.PlayerConfig.PlayerPath)
	return pm.GetFirst()
}

// SaveConfig 保存配置
func (a *App) SaveConfig() error {
	return a.PlayerConfig.Save()
}

// GetAllSources 获取所有媒体源
func (a *App) GetAllSources() []source.MediaSource {
	return a.SourceManager.GetAll()
}
