package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type loadingFinishedMsg struct {
	err error
}

// loadingPageModel 用于在耗时任务执行期间持续保持在 TUI 中。
type loadingPageModel struct {
	title    string
	subtitle string
	spinner  spinner.Model
	task     func() error
	err      error
}

// RunLoadingTUI 在后台执行任务，同时保持一个加载中的 TUI 页面。
func RunLoadingTUI(title, subtitle string, task func() error) error {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

	model := loadingPageModel{
		title:    title,
		subtitle: subtitle,
		spinner:  spin,
		task:     task,
	}

	finalModel, err := newProgram(model).Run()
	if err != nil {
		return err
	}

	result, ok := finalModel.(loadingPageModel)
	if !ok {
		return fmt.Errorf("无法读取加载页面状态")
	}
	return result.err
}

func (m loadingPageModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runTaskCmd())
}

func (m loadingPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case loadingFinishedMsg:
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.err = fmt.Errorf("已取消")
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m loadingPageModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render(m.title))
	b.WriteString("\n")
	if strings.TrimSpace(m.subtitle) != "" {
		b.WriteString(tuiMutedStyle.Render(m.subtitle))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s 正在处理，请稍候...", m.spinner.View()))
	b.WriteString("\n\n")
	b.WriteString(tuiHintStyle.Render("Ctrl+C 取消"))

	return strings.TrimRight(b.String(), "\n")
}

func (m loadingPageModel) runTaskCmd() tea.Cmd {
	return func() tea.Msg {
		return loadingFinishedMsg{err: m.task()}
	}
}
