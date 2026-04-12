package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ConfigMenuAction 表示配置页中的后续动作。
type ConfigMenuAction string

const (
	ConfigMenuActionPlayers       ConfigMenuAction = "players"
	ConfigMenuActionSubscriptions ConfigMenuAction = "subscriptions"
	ConfigMenuActionSourceChannel ConfigMenuAction = "source_channel"
	ConfigMenuActionOpenConfig    ConfigMenuAction = "open_config"
	ConfigMenuActionOpenSourceCfg ConfigMenuAction = "open_source_cfg"
	ConfigMenuActionBack          ConfigMenuAction = "back"
)

type configMenuItem struct {
	title       string
	description string
	action      ConfigMenuAction
}

// configMenuModel 管理配置页的列表选择和快捷键行为。
type configMenuModel struct {
	items    []configMenuItem
	selected int
	action   ConfigMenuAction
}

// RunConfigMenuTUI 运行配置子菜单。
func RunConfigMenuTUI() (ConfigMenuAction, error) {
	model := newConfigMenuModel()
	program := newProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return ConfigMenuActionBack, err
	}

	result, ok := finalModel.(configMenuModel)
	if !ok {
		return ConfigMenuActionBack, fmt.Errorf("无法读取配置菜单状态")
	}
	if result.action == "" {
		return ConfigMenuActionBack, nil
	}
	return result.action, nil
}

func newConfigMenuModel() configMenuModel {
	return configMenuModel{
		items: []configMenuItem{
			{
				title:       "播放器",
				description: "直接编辑播放器路径，并设置默认播放器",
				action:      ConfigMenuActionPlayers,
			},
			{
				title:       "订阅源",
				description: "直接新增、编辑、删除和刷新订阅源",
				action:      ConfigMenuActionSubscriptions,
			},
			{
				title:       "片源渠道",
				description: "按渠道开关媒体源，并查看最近一次 doctor 结果",
				action:      ConfigMenuActionSourceChannel,
			},
			{
				title:       "打开 config.json",
				description: "用系统默认编辑器打开配置文件",
				action:      ConfigMenuActionOpenConfig,
			},
			{
				title:       "打开 source_preferences.json",
				description: "打开片源渠道偏好文件，直接编辑启用状态和 doctor 结果",
				action:      ConfigMenuActionOpenSourceCfg,
			},
			{
				title:       "返回",
				description: "返回 goani TUI 首页",
				action:      ConfigMenuActionBack,
			},
		},
	}
}

func (m configMenuModel) Init() tea.Cmd {
	return nil
}

func (m configMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.action = ConfigMenuActionBack
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "enter":
			m.action = m.items[m.selected].action
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m configMenuModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render("配置"))
	b.WriteString("\n")
	b.WriteString(tuiMutedStyle.Render("查看配置说明，或直接打开 config.json"))
	b.WriteString("\n\n")

	for i, item := range m.items {
		line := fmt.Sprintf("%s %s", pointer(i == m.selected), item.title)
		if i == m.selected {
			b.WriteString(tuiPickStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(tuiMutedStyle.Render(item.description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(tuiHintStyle.Render("↑/↓ 选择，Enter 确认，Esc 返回"))

	return strings.TrimRight(b.String(), "\n")
}
