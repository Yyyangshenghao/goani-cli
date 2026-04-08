package app

import (
	"github.com/Yyyangshenghao/goani-cli/internal/config"
	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

// App 应用核心，管理全局资源和配置
type App struct {
	Config  *config.Config
	Sources []source.MediaSource
}

// New 创建应用实例
func New() *App {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	srcConfig, err := source.LoadConfig("mediaSourceJson/css1.json")
	if err != nil {
		return &App{Config: cfg, Sources: nil}
	}

	return &App{
		Config:  cfg,
		Sources: srcConfig.ExportedMediaSourceDataList.MediaSources,
	}
}

// GetSource 获取指定索引的媒体源
func (a *App) GetSource(index int) *webselector.WebSelectorSource {
	if index < 0 || index >= len(a.Sources) {
		return nil
	}
	return webselector.New(a.Sources[index])
}

// GetFirstSource 获取第一个媒体源
func (a *App) GetFirstSource() *webselector.WebSelectorSource {
	return a.GetSource(0)
}

// GetPlayer 获取配置的播放器
func (a *App) GetPlayer() player.Player {
	pm := player.NewManagerWithConfig(a.Config.PlayerName, a.Config.PlayerPath)
	return pm.GetFirst()
}

// SaveConfig 保存配置
func (a *App) SaveConfig() error {
	return a.Config.Save()
}
