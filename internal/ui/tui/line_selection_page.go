package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LineSelectionItem 描述一条已解析的候选线路。
type LineSelectionItem struct {
	Title      string
	EpisodeURL string
	VideoURL   string
	Format     string
	Quality    string
	Error      string
}

// LineSelectionResult 表示线路页最终选中的线路索引。
type LineSelectionResult struct {
	Index int
}

// lineSelectionModel 管理线路列表和当前选中线路的详情预览。
type lineSelectionModel struct {
	animeName    string
	episodeLabel string
	items        []LineSelectionItem
	selected     int
	result       *LineSelectionResult
	windowHeight int
}

// RunLineSelectionTUI 展示可播放线路列表，并在底部展示当前线路的详细信息。
func RunLineSelectionTUI(animeName, episodeLabel string, items []LineSelectionItem) (*LineSelectionResult, error) {
	model := lineSelectionModel{
		animeName:    animeName,
		episodeLabel: episodeLabel,
		items:        items,
	}

	finalModel, err := newProgram(model).Run()
	if err != nil {
		return nil, err
	}

	result, ok := finalModel.(lineSelectionModel)
	if !ok {
		return nil, fmt.Errorf("无法读取线路选择页状态")
	}

	return result.result, nil
}

func (m lineSelectionModel) Init() tea.Cmd {
	return nil
}

func (m lineSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.result = nil
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
			if len(m.items) == 0 {
				return m, nil
			}
			m.result = &LineSelectionResult{Index: m.selected}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m lineSelectionModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render("线路选择"))
	b.WriteString("\n")
	b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("番剧: %s    剧集: %s", m.animeName, m.episodeLabel)))
	b.WriteString("\n\n")

	if len(m.items) == 0 {
		b.WriteString(tuiMutedStyle.Render("当前剧集没有可展示的线路"))
		b.WriteString("\n")
		b.WriteString(tuiHintStyle.Render("Esc 返回"))
		return strings.TrimRight(b.String(), "\n")
	}

	b.WriteString(tuiTitleStyle.Render("可选线路"))
	b.WriteString("\n")

	start, end := m.visibleRange()
	for i := start; i < end; i++ {
		item := m.items[i]
		status := "可播放"
		if strings.TrimSpace(item.Error) != "" {
			status = "解析失败"
		}

		line := fmt.Sprintf("%s %d. %s  [%s / %s / %s]", pointer(i == m.selected), i+1, item.Title, displayOrUnknown(item.Format), displayOrUnknown(item.Quality), status)
		if i == m.selected {
			b.WriteString(tuiPickStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	if selected := m.currentItem(); selected != nil {
		b.WriteString("\n")
		b.WriteString(tuiTitleStyle.Render("当前线路详情"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("线路名: %s\n", selected.Title))
		b.WriteString(fmt.Sprintf("剧集页: %s\n", displayOrUnknown(selected.EpisodeURL)))
		b.WriteString(fmt.Sprintf("视频链: %s\n", displayOrUnknown(selected.VideoURL)))
		b.WriteString(fmt.Sprintf("格式: %s\n", displayOrUnknown(selected.Format)))
		b.WriteString(fmt.Sprintf("清晰度: %s\n", displayOrUnknown(selected.Quality)))
		if strings.TrimSpace(selected.Error) != "" {
			b.WriteString(tuiErrorStyle.Render("错误: " + selected.Error))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(tuiHintStyle.Render("↑/↓ 选择线路，Enter 播放当前线路，Esc 返回选集"))

	return strings.TrimRight(b.String(), "\n")
}

func (m lineSelectionModel) currentItem() *LineSelectionItem {
	if len(m.items) == 0 || m.selected < 0 || m.selected >= len(m.items) {
		return nil
	}
	return &m.items[m.selected]
}

func (m lineSelectionModel) visibleRange() (int, int) {
	page := m.pageSize()
	start := m.selected - page/2
	if start < 0 {
		start = 0
	}
	end := start + page
	if end > len(m.items) {
		end = len(m.items)
		start = end - page
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func (m lineSelectionModel) pageSize() int {
	if m.windowHeight <= 18 {
		return 6
	}
	return m.windowHeight / 3
}

func displayOrUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "未知"
	}
	return value
}
