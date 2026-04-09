package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SubscriptionConfigPageItem 描述订阅源配置页中的一条订阅记录。
type SubscriptionConfigPageItem struct {
	Name      string
	URL       string
	UpdatedAt string
}

// SubscriptionConfigPageState 是订阅源配置页的完整渲染状态。
type SubscriptionConfigPageState struct {
	ConfigPath  string
	SourceCount int
	Items       []SubscriptionConfigPageItem
}

// SaveSubscriptionFunc 保存新增或编辑后的订阅；editingURL 为空表示新增。
type SaveSubscriptionFunc func(editingURL, name, url string) (SubscriptionConfigPageState, error)

// RemoveSubscriptionFunc 删除指定 URL 的订阅。
type RemoveSubscriptionFunc func(url string) (SubscriptionConfigPageState, error)

// RefreshSubscriptionsFunc 刷新当前所有订阅。
type RefreshSubscriptionsFunc func() (SubscriptionConfigPageState, error)

// ResetSubscriptionsFunc 把订阅恢复为默认配置。
type ResetSubscriptionsFunc func() (SubscriptionConfigPageState, error)

// subscriptionConfigPageModel 管理订阅列表浏览、编辑和刷新操作。
type subscriptionConfigPageModel struct {
	state      SubscriptionConfigPageState
	save       SaveSubscriptionFunc
	remove     RemoveSubscriptionFunc
	refresh    RefreshSubscriptionsFunc
	reset      ResetSubscriptionsFunc
	selected   int
	editing    bool
	editingURL string
	focusField int
	nameInput  textinput.Model
	urlInput   textinput.Model
	status     string
	statusErr  bool
}

// RunSubscriptionConfigTUI 运行订阅源配置页。
func RunSubscriptionConfigTUI(
	initial SubscriptionConfigPageState,
	save SaveSubscriptionFunc,
	remove RemoveSubscriptionFunc,
	refresh RefreshSubscriptionsFunc,
	reset ResetSubscriptionsFunc,
) error {
	model := newSubscriptionConfigPageModel(initial, save, remove, refresh, reset)
	_, err := newProgram(model).Run()
	return err
}

func newSubscriptionConfigPageModel(
	initial SubscriptionConfigPageState,
	save SaveSubscriptionFunc,
	remove RemoveSubscriptionFunc,
	refresh RefreshSubscriptionsFunc,
	reset ResetSubscriptionsFunc,
) subscriptionConfigPageModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "输入订阅名称（可选）"
	nameInput.CharLimit = 128
	nameInput.Width = 44

	urlInput := textinput.New()
	urlInput.Placeholder = "输入订阅 URL"
	urlInput.CharLimit = 512
	urlInput.Width = 72

	return subscriptionConfigPageModel{
		state:     initial,
		save:      save,
		remove:    remove,
		refresh:   refresh,
		reset:     reset,
		nameInput: nameInput,
		urlInput:  urlInput,
	}
}

func (m subscriptionConfigPageModel) Init() tea.Cmd {
	return nil
}

func (m subscriptionConfigPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateBrowsing(msg)
	}

	return m, nil
}

func (m subscriptionConfigPageModel) updateBrowsing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "a":
		m.beginEditing("", "", "")
		m.status = "正在新增订阅"
		m.statusErr = false
		return m, textinput.Blink
	case "enter":
		if len(m.state.Items) == 0 {
			return m, nil
		}
		item := m.state.Items[m.selected]
		m.beginEditing(item.URL, item.Name, item.URL)
		m.status = fmt.Sprintf("正在编辑订阅: %s", item.Name)
		m.statusErr = false
		return m, textinput.Blink
	case "d":
		if len(m.state.Items) == 0 {
			return m, nil
		}
		item := m.state.Items[m.selected]
		nextState, err := m.remove(item.URL)
		if err != nil {
			m.status = err.Error()
			m.statusErr = true
			return m, nil
		}
		m.state = nextState
		m.restoreSelection(item.URL)
		m.status = fmt.Sprintf("已移除订阅: %s", item.Name)
		m.statusErr = false
	case "r":
		nextState, err := m.refresh()
		if err != nil {
			m.status = err.Error()
			m.statusErr = true
			return m, nil
		}
		m.state = nextState
		m.status = "已刷新订阅"
		m.statusErr = false
	case "z":
		nextState, err := m.reset()
		if err != nil {
			m.status = err.Error()
			m.statusErr = true
			return m, nil
		}
		m.state = nextState
		m.selected = 0
		m.status = "已重置为默认订阅"
		m.statusErr = false
	}

	return m, nil
}

func (m subscriptionConfigPageModel) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.endEditing("已取消编辑", false)
		return m, nil
	case "tab", "shift+tab", "up", "down":
		m.switchFocus()
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		url := strings.TrimSpace(m.urlInput.Value())
		if url == "" {
			m.status = "订阅 URL 不能为空"
			m.statusErr = true
			return m, nil
		}
		nextState, err := m.save(m.editingURL, name, url)
		if err != nil {
			m.status = err.Error()
			m.statusErr = true
			return m, nil
		}
		m.state = nextState
		restoreKey := url
		if m.editingURL != "" {
			restoreKey = url
		}
		m.restoreSelection(restoreKey)
		m.endEditing("已保存订阅", false)
		return m, nil
	}

	var cmd tea.Cmd
	if m.focusField == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.urlInput, cmd = m.urlInput.Update(msg)
	}
	return m, cmd
}

func (m subscriptionConfigPageModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render("订阅源配置"))
	b.WriteString("\n")
	b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("订阅数: %d    已加载媒体源: %d    配置文件: %s", len(m.state.Items), m.state.SourceCount, m.state.ConfigPath)))
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
	b.WriteString(tuiTitleStyle.Render("当前订阅"))
	b.WriteString("\n")

	if len(m.state.Items) == 0 {
		b.WriteString(tuiMutedStyle.Render("暂无订阅，按 a 新增，或按 z 恢复默认订阅"))
		b.WriteString("\n")
	} else {
		for i, item := range m.state.Items {
			line := fmt.Sprintf("%s %s", pointer(i == m.selected), displaySubscriptionName(item.Name))
			if i == m.selected {
				b.WriteString(tuiPickStyle.Render(line))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
			b.WriteString("  URL: ")
			b.WriteString(item.URL)
			b.WriteString("\n")
			if strings.TrimSpace(item.UpdatedAt) != "" {
				b.WriteString("  更新时间: ")
				b.WriteString(item.UpdatedAt)
				b.WriteString("\n")
			}
		}
	}

	if m.editing {
		b.WriteString("\n")
		b.WriteString(tuiTitleStyle.Render("编辑订阅"))
		b.WriteString("\n")
		b.WriteString(m.renderEditField("名称", m.nameInput, m.focusField == 0))
		b.WriteString("\n")
		b.WriteString(m.renderEditField("URL", m.urlInput, m.focusField == 1))
		b.WriteString("\n")
		b.WriteString(tuiHintStyle.Render("Tab 切换输入框，Enter 保存，Esc 取消"))
	} else {
		b.WriteString("\n")
		b.WriteString(tuiHintStyle.Render("↑/↓ 选择，Enter 编辑，a 新增，d 删除，r 刷新，z 重置默认，Esc 返回"))
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m *subscriptionConfigPageModel) beginEditing(editingURL, name, url string) {
	m.editing = true
	m.editingURL = editingURL
	m.focusField = 0
	m.nameInput.SetValue(name)
	m.nameInput.Focus()
	m.urlInput.SetValue(url)
	m.urlInput.Blur()
}

func (m *subscriptionConfigPageModel) endEditing(status string, isErr bool) {
	m.editing = false
	m.editingURL = ""
	m.focusField = 0
	m.nameInput.Blur()
	m.urlInput.Blur()
	m.status = status
	m.statusErr = isErr
}

func (m *subscriptionConfigPageModel) switchFocus() {
	if m.focusField == 0 {
		m.focusField = 1
		m.nameInput.Blur()
		m.urlInput.Focus()
		return
	}
	m.focusField = 0
	m.urlInput.Blur()
	m.nameInput.Focus()
}

func (m subscriptionConfigPageModel) renderEditField(label string, input textinput.Model, focused bool) string {
	if focused {
		return tuiPickStyle.Render(label) + "\n" + input.View()
	}
	return label + "\n" + input.View()
}

func (m *subscriptionConfigPageModel) restoreSelection(url string) {
	for i, item := range m.state.Items {
		if item.URL == url {
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

func displaySubscriptionName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "未命名订阅"
	}
	return name
}
