package workflow

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/settings"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

// RunConfigTUI 负责配置页循环、状态构建和保存动作编排。
func RunConfigTUI(application *app.App) error {
	for {
		action, err := tui.RunConfigMenuTUI()
		if err != nil {
			return err
		}

		switch action {
		case tui.ConfigMenuActionPlayers:
			if err := runPlayerConfigTUI(application); err != nil {
				return err
			}
		case tui.ConfigMenuActionSubscriptions:
			if err := runSubscriptionConfigTUI(application); err != nil {
				return err
			}
		case tui.ConfigMenuActionOpenConfig:
			if err := openConfigInDefaultEditor(application); err != nil {
				if pageErr := tui.RunTextTUI("打开配置失败", err.Error()); pageErr != nil {
					return fmt.Errorf("%w；另外打开失败页面也出错: %v", err, pageErr)
				}
				continue
			}

			configPath, pathErr := settings.GetConfigPath()
			if pathErr != nil {
				configPath = "config.json"
			}
			if err := tui.RunTextTUI("已打开配置文件", fmt.Sprintf("已尝试用系统默认编辑器打开:\n%s\n\n如果没有弹出编辑器，请检查系统对 .json 文件的默认打开方式。", configPath)); err != nil {
				return err
			}
		case tui.ConfigMenuActionBack:
			return nil
		default:
			return nil
		}
	}
}

// RenderSourceOverview 返回 TUI 首页“媒体源”页面的文本内容。
func RenderSourceOverview(application *app.App) string {
	var b strings.Builder

	sources := application.SourceManager.GetAll()
	subs := application.SourceManager.GetSubscriptions()

	b.WriteString(fmt.Sprintf("已加载媒体源: %d\n", len(sources)))
	b.WriteString("\n")
	for i, s := range sources {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, s.Arguments.Name))
		if strings.TrimSpace(s.Arguments.Description) != "" {
			b.WriteString(fmt.Sprintf("   %s\n", s.Arguments.Description))
		}
	}

	b.WriteString("\n订阅列表:\n")
	if len(subs) == 0 {
		b.WriteString("暂无订阅\n")
		return b.String()
	}

	for i, sub := range subs {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, sub.Name))
		b.WriteString(fmt.Sprintf("   URL: %s\n", sub.URL))
		if strings.TrimSpace(sub.UpdatedAt) != "" {
			b.WriteString(fmt.Sprintf("   更新时间: %s\n", sub.UpdatedAt))
		}
	}

	return b.String()
}

func runPlayerConfigTUI(application *app.App) error {
	return tui.RunPlayerConfigTUI(
		buildPlayerConfigPageState(application),
		func(name, path string) (tui.PlayerConfigPageState, error) {
			path = strings.TrimSpace(path)
			if path == "" {
				return tui.PlayerConfigPageState{}, fmt.Errorf("路径不能为空")
			}
			if !player.IsSupportedPlayer(name) {
				return tui.PlayerConfigPageState{}, fmt.Errorf("不支持的播放器: %s", name)
			}

			application.PlayerConfig.SetPlayerPath(name, path)
			if err := application.SaveConfig(); err != nil {
				return tui.PlayerConfigPageState{}, fmt.Errorf("保存配置失败: %w", err)
			}
			return buildPlayerConfigPageState(application), nil
		},
		func(name string) (tui.PlayerConfigPageState, error) {
			if strings.TrimSpace(application.PlayerConfig.GetPath(name)) == "" {
				return tui.PlayerConfigPageState{}, fmt.Errorf("请先为 %s 配置路径", name)
			}

			application.PlayerConfig.SetDefaultPlayer(name)
			if err := application.SaveConfig(); err != nil {
				return tui.PlayerConfigPageState{}, fmt.Errorf("保存配置失败: %w", err)
			}
			return buildPlayerConfigPageState(application), nil
		},
	)
}

func runSubscriptionConfigTUI(application *app.App) error {
	return tui.RunSubscriptionConfigTUI(
		buildSubscriptionConfigPageState(application),
		func(editingURL, name, url string) (tui.SubscriptionConfigPageState, error) {
			url = strings.TrimSpace(url)
			name = strings.TrimSpace(name)
			if url == "" {
				return tui.SubscriptionConfigPageState{}, fmt.Errorf("订阅 URL 不能为空")
			}
			if err := application.SourceManager.UpsertSubscription(editingURL, url, name); err != nil {
				return tui.SubscriptionConfigPageState{}, err
			}
			return buildSubscriptionConfigPageState(application), nil
		},
		func(url string) (tui.SubscriptionConfigPageState, error) {
			if err := application.SourceManager.Unsubscribe(url); err != nil {
				return tui.SubscriptionConfigPageState{}, err
			}
			return buildSubscriptionConfigPageState(application), nil
		},
		func() (tui.SubscriptionConfigPageState, error) {
			if err := application.SourceManager.Refresh(); err != nil {
				return tui.SubscriptionConfigPageState{}, err
			}
			return buildSubscriptionConfigPageState(application), nil
		},
		func() (tui.SubscriptionConfigPageState, error) {
			if err := application.SourceManager.Reset(); err != nil {
				return tui.SubscriptionConfigPageState{}, err
			}
			return buildSubscriptionConfigPageState(application), nil
		},
	)
}

func buildPlayerConfigPageState(application *app.App) tui.PlayerConfigPageState {
	configPath, err := settings.GetConfigPath()
	if err != nil {
		configPath = "~/.goani/config.json"
	}

	supported := player.SupportedPlayers()
	configured := make([]tui.PlayerConfigPageItem, 0, len(supported))
	unconfigured := make([]tui.PlayerConfigPageItem, 0, len(supported))

	for _, name := range supported {
		item := tui.PlayerConfigPageItem{
			Name:       name,
			Path:       application.PlayerConfig.GetPath(name),
			Configured: strings.TrimSpace(application.PlayerConfig.GetPath(name)) != "",
			IsDefault:  application.PlayerConfig.DefaultPlayer == name,
		}
		if item.Configured {
			configured = append(configured, item)
		} else {
			unconfigured = append(unconfigured, item)
		}
	}

	return tui.PlayerConfigPageState{
		ConfigPath: configPath,
		Items:      append(configured, unconfigured...),
	}
}

func buildSubscriptionConfigPageState(application *app.App) tui.SubscriptionConfigPageState {
	configPath, err := settings.GetConfigPath()
	if err != nil {
		configPath = "~/.goani/config.json"
	}

	subs := application.SourceManager.GetSubscriptions()
	items := make([]tui.SubscriptionConfigPageItem, 0, len(subs))
	for _, sub := range subs {
		items = append(items, tui.SubscriptionConfigPageItem{
			Name:      sub.Name,
			URL:       sub.URL,
			UpdatedAt: sub.UpdatedAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return tui.SubscriptionConfigPageState{
		ConfigPath:  configPath,
		SourceCount: application.SourceManager.Count(),
		Items:       items,
	}
}

func openConfigInDefaultEditor(application *app.App) error {
	if err := application.Settings.Save(); err != nil {
		return fmt.Errorf("创建配置文件失败: %w", err)
	}

	configPath, err := settings.GetConfigPath()
	if err != nil {
		return fmt.Errorf("获取配置文件路径失败: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", configPath)
	case "darwin":
		cmd = exec.Command("open", configPath)
	default:
		cmd = exec.Command("xdg-open", configPath)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("打开配置文件失败: %w", err)
	}

	return nil
}
