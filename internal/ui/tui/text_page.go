package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type textTUIModel struct {
	title  string
	lines  []string
	width  int
	height int
	offset int
}

// RunTextTUI 运行只读信息页面，支持滚动和返回
func RunTextTUI(title, content string) error {
	model := textTUIModel{
		title: title,
		lines: strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n"),
	}
	program := newProgram(model)
	_, err := program.Run()
	return err
}

func (m textTUIModel) Init() tea.Cmd {
	return nil
}

func (m textTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q", "enter":
			return m, tea.Quit
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			if m.offset < m.maxOffset() {
				m.offset++
			}
		case "pgdown":
			m.offset += m.pageSize()
			if m.offset > m.maxOffset() {
				m.offset = m.maxOffset()
			}
		case "pgup":
			m.offset -= m.pageSize()
			if m.offset < 0 {
				m.offset = 0
			}
		case "home":
			m.offset = 0
		case "end":
			m.offset = m.maxOffset()
		}
	}

	return m, nil
}

func (m textTUIModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render(m.title))
	b.WriteString("\n")
	b.WriteString(tuiHintStyle.Render("Esc 返回，↑/↓ 滚动，PgUp/PgDn 翻页"))
	b.WriteString("\n\n")

	start := m.offset
	end := start + m.pageSize()
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for i := start; i < end; i++ {
		b.WriteString(m.lines[i])
		b.WriteString("\n")
	}

	if len(m.lines) == 0 {
		b.WriteString(tuiMutedStyle.Render("没有可显示的内容"))
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m textTUIModel) pageSize() int {
	if m.height <= 6 {
		return 12
	}
	return m.height - 4
}

func (m textTUIModel) maxOffset() int {
	max := len(m.lines) - m.pageSize()
	if max < 0 {
		return 0
	}
	return max
}
