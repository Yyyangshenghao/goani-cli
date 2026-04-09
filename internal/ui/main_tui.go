package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MainMenuAction 主菜单动作
type MainMenuAction string

const (
	MainMenuActionSearch     MainMenuAction = "search"
	MainMenuActionSourceList MainMenuAction = "source_list"
	MainMenuActionConfigHelp MainMenuAction = "config_help"
	MainMenuActionVersion    MainMenuAction = "version"
	MainMenuActionQuit       MainMenuAction = "quit"
)

type mainMenuItem struct {
	title       string
	description string
	action      MainMenuAction
}

type mainTUIModel struct {
	items    []mainMenuItem
	selected int
	action   MainMenuAction
}

// RunMainTUI 运行主入口菜单
func RunMainTUI() (MainMenuAction, error) {
	model := newMainTUIModel()
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return MainMenuActionQuit, err
	}

	result, ok := finalModel.(mainTUIModel)
	if !ok {
		return MainMenuActionQuit, fmt.Errorf("无法读取主菜单状态")
	}
	if result.action == "" {
		return MainMenuActionQuit, nil
	}
	return result.action, nil
}

func newMainTUIModel() mainTUIModel {
	return mainTUIModel{
		items: []mainMenuItem{
			{
				title:       "实时搜索",
				description: "进入交互式搜索 TUI，搜索、选源、选番剧",
				action:      MainMenuActionSearch,
			},
			{
				title:       "媒体源",
				description: "查看当前已加载的片源和订阅列表",
				action:      MainMenuActionSourceList,
			},
			{
				title:       "配置说明",
				description: "查看播放器配置和 config.json 的说明",
				action:      MainMenuActionConfigHelp,
			},
			{
				title:       "版本信息",
				description: "查看当前程序版本",
				action:      MainMenuActionVersion,
			},
			{
				title:       "退出",
				description: "退出 goani",
				action:      MainMenuActionQuit,
			},
		},
	}
}

func (m mainTUIModel) Init() tea.Cmd {
	return nil
}

func (m mainTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.action = MainMenuActionQuit
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

func (m mainTUIModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render("goani TUI"))
	b.WriteString("\n")
	b.WriteString(tuiMutedStyle.Render("选择一个操作开始，Esc 或 q 退出"))
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
	b.WriteString(tuiHintStyle.Render("↑/↓ 选择，Enter 确认，Esc 退出"))

	return strings.TrimRight(b.String(), "\n")
}
