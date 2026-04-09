package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// PlayerConfigPageItem 描述播放器配置页里的一条播放器记录。
type PlayerConfigPageItem struct {
	Name       string
	Path       string
	Configured bool
	IsDefault  bool
}

// PlayerConfigPageState 是播放器配置页渲染所需的完整状态快照。
type PlayerConfigPageState struct {
	ConfigPath string
	Items      []PlayerConfigPageItem
}

// SavePlayerPathFunc 负责把页面里输入的播放器路径保存到真实配置。
type SavePlayerPathFunc func(name, path string) (PlayerConfigPageState, error)

// SetDefaultPlayerFunc 负责把指定播放器设为默认播放器。
type SetDefaultPlayerFunc func(name string) (PlayerConfigPageState, error)

// playerConfigPageModel 管理播放器配置页的列表浏览和路径编辑状态。
type playerConfigPageModel struct {
	state      PlayerConfigPageState
	savePath   SavePlayerPathFunc
	setDefault SetDefaultPlayerFunc
	selected   int
	editing    bool
	input      textinput.Model
	status     string
	statusErr  bool
}

// RunPlayerConfigTUI 运行可直接编辑播放器路径的配置页。
func RunPlayerConfigTUI(initial PlayerConfigPageState, savePath SavePlayerPathFunc, setDefault SetDefaultPlayerFunc) error {
	model := newPlayerConfigPageModel(initial, savePath, setDefault)
	_, err := newProgram(model).Run()
	return err
}

func newPlayerConfigPageModel(initial PlayerConfigPageState, savePath SavePlayerPathFunc, setDefault SetDefaultPlayerFunc) playerConfigPageModel {
	input := textinput.New()
	input.Placeholder = "输入播放器路径"
	input.CharLimit = 512
	input.Width = 72

	return playerConfigPageModel{
		state:      initial,
		savePath:   savePath,
		setDefault: setDefault,
		input:      input,
	}
}

func (m playerConfigPageModel) Init() tea.Cmd {
	return nil
}

func (m playerConfigPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateBrowsing(msg)
	}

	return m, nil
}

func (m playerConfigPageModel) updateBrowsing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		return m, tea.Quit
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.state.Items)-1 {
			m.selected++
		}
	case "enter":
		if len(m.state.Items) == 0 {
			return m, nil
		}
		item := m.state.Items[m.selected]
		m.editing = true
		m.input.SetValue(item.Path)
		m.input.CursorEnd()
		m.input.Focus()
		m.status = fmt.Sprintf("正在编辑 %s 的路径", item.Name)
		m.statusErr = false
		return m, textinput.Blink
	case "d":
		if len(m.state.Items) == 0 {
			return m, nil
		}
		item := m.state.Items[m.selected]
		if !item.Configured {
			m.status = fmt.Sprintf("%s 还没有配置路径，暂时不能设为默认", item.Name)
			m.statusErr = true
			return m, nil
		}

		nextState, err := m.setDefault(item.Name)
		if err != nil {
			m.status = err.Error()
			m.statusErr = true
			return m, nil
		}

		m.state = nextState
		m.restoreSelection(item.Name)
		m.status = fmt.Sprintf("已将 %s 设为默认播放器", item.Name)
		m.statusErr = false
	}

	return m, nil
}

func (m playerConfigPageModel) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.input.Blur()
		m.status = "已取消编辑"
		m.statusErr = false
		return m, nil
	case "enter":
		if len(m.state.Items) == 0 {
			return m, nil
		}

		item := m.state.Items[m.selected]
		path := strings.TrimSpace(m.input.Value())
		if path == "" {
			m.status = "路径不能为空"
			m.statusErr = true
			return m, nil
		}

		nextState, err := m.savePath(item.Name, path)
		if err != nil {
			m.status = err.Error()
			m.statusErr = true
			return m, nil
		}

		m.state = nextState
		m.editing = false
		m.input.Blur()
		m.restoreSelection(item.Name)
		m.status = fmt.Sprintf("已保存 %s 的路径", item.Name)
		m.statusErr = false
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m playerConfigPageModel) View() string {
	var b strings.Builder

	configuredCount := 0
	defaultPlayer := "未设置"
	for _, item := range m.state.Items {
		if item.Configured {
			configuredCount++
		}
		if item.IsDefault {
			defaultPlayer = item.Name
		}
	}

	b.WriteString(tuiTitleStyle.Render("播放器配置"))
	b.WriteString("\n")
	b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("默认播放器: %s    已配置: %d    配置文件: %s", defaultPlayer, configuredCount, m.state.ConfigPath)))
	b.WriteString("\n")

	if m.status != "" {
		if m.statusErr {
			b.WriteString(tuiErrorStyle.Render(m.status))
		} else {
			b.WriteString(tuiOkStyle.Render(m.status))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(tuiTitleStyle.Render("已配置"))
	b.WriteString("\n")

	hasConfigured := false
	for i, item := range m.state.Items {
		if !item.Configured {
			continue
		}
		hasConfigured = true
		b.WriteString(m.renderItem(i, item))
	}
	if !hasConfigured {
		b.WriteString(tuiMutedStyle.Render("暂无已配置播放器"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(tuiTitleStyle.Render("可配置播放器"))
	b.WriteString("\n")
	hasUnconfigured := false
	for i, item := range m.state.Items {
		if item.Configured {
			continue
		}
		hasUnconfigured = true
		b.WriteString(m.renderItem(i, item))
	}
	if !hasUnconfigured {
		b.WriteString(tuiMutedStyle.Render("所有内置播放器都已经配置过路径"))
		b.WriteString("\n")
	}

	if m.editing && len(m.state.Items) > 0 {
		item := m.state.Items[m.selected]
		b.WriteString("\n")
		b.WriteString(tuiTitleStyle.Render("编辑路径"))
		b.WriteString("\n")
		b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("正在编辑: %s", item.Name)))
		b.WriteString("\n")
		b.WriteString(m.input.View())
		b.WriteString("\n")
		b.WriteString(tuiHintStyle.Render("Enter 保存，Esc 取消"))
	} else {
		b.WriteString("\n")
		b.WriteString(tuiHintStyle.Render("↑/↓ 选择，Enter 编辑路径，d 设为默认，Esc 返回"))
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m playerConfigPageModel) renderItem(index int, item PlayerConfigPageItem) string {
	var b strings.Builder

	flags := make([]string, 0, 2)
	if item.IsDefault {
		flags = append(flags, "默认")
	}
	if item.Configured {
		flags = append(flags, "已配置")
	} else {
		flags = append(flags, "未配置")
	}

	line := fmt.Sprintf("%s %s [%s]", pointer(index == m.selected), item.Name, strings.Join(flags, " / "))
	if index == m.selected {
		b.WriteString(tuiPickStyle.Render(line))
	} else {
		b.WriteString(line)
	}
	b.WriteString("\n")
	if item.Configured {
		b.WriteString("  路径: ")
		b.WriteString(item.Path)
	} else {
		b.WriteString(tuiMutedStyle.Render("  路径: 未设置"))
	}
	b.WriteString("\n")

	return b.String()
}

func (m *playerConfigPageModel) restoreSelection(name string) {
	for i, item := range m.state.Items {
		if item.Name == name {
			m.selected = i
			return
		}
	}
	if len(m.state.Items) == 0 {
		m.selected = 0
		return
	}
	if m.selected >= len(m.state.Items) {
		m.selected = len(m.state.Items) - 1
	}
}
