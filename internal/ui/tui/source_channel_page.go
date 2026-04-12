package tui

import (
	"fmt"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
	tea "github.com/charmbracelet/bubbletea"
)

// SourceChannelPageItem 描述片源渠道配置页中的一条记录。
type SourceChannelPageItem struct {
	ID          string
	Name        string
	Description string
	Enabled     bool
	Priority    int
	LastDoctor  *source.SourceDoctorSnapshot
}

// SourceChannelPageState 是片源渠道配置页的完整渲染状态。
type SourceChannelPageState struct {
	ConfigPath   string
	TotalCount   int
	EnabledCount int
	Items        []SourceChannelPageItem
}

// ToggleSourceChannelFunc 切换单个渠道的启用状态。
type ToggleSourceChannelFunc func(id string, enabled bool) (SourceChannelPageState, error)

// EnableAllSourceChannelsFunc 启用全部渠道。
type EnableAllSourceChannelsFunc func() (SourceChannelPageState, error)

// EnableLastWorkingSourceChannelsFunc 仅启用最近诊断成功的渠道。
type EnableLastWorkingSourceChannelsFunc func() (SourceChannelPageState, error)

type sourceChannelPageModel struct {
	state             SourceChannelPageState
	toggle            ToggleSourceChannelFunc
	enableAll         EnableAllSourceChannelsFunc
	enableLastWorking EnableLastWorkingSourceChannelsFunc
	selected          int
	status            string
	statusErr         bool
}

// RunSourceChannelConfigTUI 运行片源渠道配置页。
func RunSourceChannelConfigTUI(
	initial SourceChannelPageState,
	toggle ToggleSourceChannelFunc,
	enableAll EnableAllSourceChannelsFunc,
	enableLastWorking EnableLastWorkingSourceChannelsFunc,
) error {
	model := sourceChannelPageModel{
		state:             initial,
		toggle:            toggle,
		enableAll:         enableAll,
		enableLastWorking: enableLastWorking,
	}
	_, err := newProgram(model).Run()
	return err
}

func (m sourceChannelPageModel) Init() tea.Cmd {
	return nil
}

func (m sourceChannelPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		case " ", "enter":
			if len(m.state.Items) == 0 {
				return m, nil
			}
			item := m.state.Items[m.selected]
			nextState, err := m.toggle(item.ID, !item.Enabled)
			if err != nil {
				m.status = err.Error()
				m.statusErr = true
				return m, nil
			}
			m.state = nextState
			m.restoreSelection(item.ID)
			if item.Enabled {
				m.status = fmt.Sprintf("已禁用: %s", item.Name)
			} else {
				m.status = fmt.Sprintf("已启用: %s", item.Name)
			}
			m.statusErr = false
		case "e":
			nextState, err := m.enableAll()
			if err != nil {
				m.status = err.Error()
				m.statusErr = true
				return m, nil
			}
			m.state = nextState
			m.status = "已启用全部渠道"
			m.statusErr = false
		case "g":
			nextState, err := m.enableLastWorking()
			if err != nil {
				m.status = err.Error()
				m.statusErr = true
				return m, nil
			}
			m.state = nextState
			m.status = "已仅保留最近探测成功的渠道"
			m.statusErr = false
		}
	}

	return m, nil
}

func (m sourceChannelPageModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render("片源渠道"))
	b.WriteString("\n")
	b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("已启用: %d / %d    偏好文件: %s", m.state.EnabledCount, m.state.TotalCount, m.state.ConfigPath)))
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
	if len(m.state.Items) == 0 {
		b.WriteString(tuiMutedStyle.Render("当前还没有已加载的片源渠道"))
		b.WriteString("\n")
	} else {
		for i, item := range m.state.Items {
			b.WriteString(m.renderItem(i, item))
		}
	}

	b.WriteString("\n")
	b.WriteString(tuiHintStyle.Render("↑/↓ 选择，Enter/空格 开关，e 全开，g 仅保留最近诊断成功，Esc 返回"))
	b.WriteString("\n")
	b.WriteString(tuiMutedStyle.Render("提示: 先运行 goani source doctor，这里就能看到最近诊断结果和优先级"))

	return strings.TrimRight(b.String(), "\n")
}

func (m sourceChannelPageModel) renderItem(index int, item SourceChannelPageItem) string {
	var b strings.Builder

	statusParts := []string{channelEnabledLabel(item.Enabled)}
	if item.Priority > 0 {
		statusParts = append(statusParts, fmt.Sprintf("优先级 %d", item.Priority))
	}
	if item.LastDoctor != nil {
		statusParts = append(statusParts, fmt.Sprintf("%d/%d 成功", item.LastDoctor.SuccessfulRuns, item.LastDoctor.TotalRuns))
	}

	line := fmt.Sprintf("%s %s [%s]", pointer(index == m.selected), item.Name, strings.Join(statusParts, " / "))
	if index == m.selected {
		b.WriteString(tuiPickStyle.Render(line))
	} else {
		b.WriteString(line)
	}
	b.WriteString("\n")

	if strings.TrimSpace(item.Description) != "" {
		b.WriteString("  ")
		b.WriteString(item.Description)
		b.WriteString("\n")
	}

	if item.LastDoctor != nil {
		b.WriteString("  最近诊断: ")
		b.WriteString(item.LastDoctor.CheckedAt)
		b.WriteString("\n")
		if strings.TrimSpace(item.LastDoctor.Summary) != "" {
			b.WriteString("  ")
			b.WriteString(item.LastDoctor.Summary)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m *sourceChannelPageModel) restoreSelection(id string) {
	for i, item := range m.state.Items {
		if item.ID == id {
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

func channelEnabledLabel(enabled bool) string {
	if enabled {
		return "已启用"
	}
	return "已禁用"
}
