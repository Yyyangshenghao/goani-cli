package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
)

type selectorTUIResult struct {
	index int
}

type selectorTUIModel struct {
	title        string
	subtitle     string
	emptyMessage string
	items        []string
	jumpValues   []string
	allowReverse bool
	allowNumber  bool
	reversed     bool
	selected     int
	numberBuffer string
	result       *selectorTUIResult
	windowHeight int
}

// RunAnimeSelectionTUI 运行番剧选择 TUI
func RunAnimeSelectionTUI(sourceName string, animes []source.Anime) (*source.Anime, error) {
	items := make([]string, len(animes))
	jumpValues := make([]string, len(animes))
	for i, anime := range animes {
		items[i] = anime.Name
		jumpValues[i] = strconv.Itoa(i + 1)
	}

	model := selectorTUIModel{
		title:        fmt.Sprintf("片源: %s", sourceName),
		subtitle:     "选择番剧，Esc 返回片源列表",
		emptyMessage: "当前片源没有可选番剧",
		items:        items,
		jumpValues:   jumpValues,
		allowNumber:  true,
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, err
	}

	return &animes[result.index], nil
}

// RunEpisodeSelectionTUI 运行剧集选择 TUI
func RunEpisodeSelectionTUI(animeName string, episodes []source.EpisodeGroup) (*source.EpisodeGroup, error) {
	items := make([]string, len(episodes))
	jumpValues := make([]string, len(episodes))
	for i, episode := range episodes {
		items[i] = episode.Label()
		if episode.HasNumber && episode.Number != "" {
			jumpValues[i] = normalizeDigitInput(episode.Number)
		} else {
			jumpValues[i] = strconv.Itoa(i + 1)
		}
	}

	model := selectorTUIModel{
		title:        fmt.Sprintf("剧集: %s", animeName),
		subtitle:     "Enter 播放，r 切换顺序/倒序，输入集数跳转，Esc 返回番剧列表",
		emptyMessage: "当前番剧没有可选剧集",
		items:        items,
		jumpValues:   jumpValues,
		allowReverse: true,
		allowNumber:  true,
		reversed:     true,
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, err
	}

	return &episodes[result.index], nil
}

func runSelectorTUI(model selectorTUIModel) (*selectorTUIResult, error) {
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	resultModel, ok := finalModel.(selectorTUIModel)
	if !ok {
		return nil, fmt.Errorf("无法读取选择器状态")
	}

	return resultModel.result, nil
}

func (m selectorTUIModel) Init() tea.Cmd {
	return nil
}

func (m selectorTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m, nil
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
			return m, nil
		case "home":
			m.selected = 0
			return m, nil
		case "end":
			if len(m.items) > 0 {
				m.selected = len(m.items) - 1
			}
			return m, nil
		case "r":
			if m.allowReverse {
				actual := m.displayIndexToActual(m.selected)
				m.reversed = !m.reversed
				m.selected = m.actualToDisplayIndex(actual)
			}
			return m, nil
		case "backspace":
			if m.allowNumber && m.numberBuffer != "" {
				m.numberBuffer = m.numberBuffer[:len(m.numberBuffer)-1]
				m.applyNumberBuffer()
			}
			return m, nil
		case "enter":
			if len(m.items) == 0 {
				return m, nil
			}
			if m.allowNumber && m.numberBuffer != "" {
				m.applyNumberBuffer()
			}
			m.result = &selectorTUIResult{index: m.displayIndexToActual(m.selected)}
			return m, tea.Quit
		}

		if m.allowNumber && isDigits(msg.String()) {
			m.numberBuffer += msg.String()
			m.applyNumberBuffer()
			return m, nil
		}
	}

	return m, nil
}

func (m selectorTUIModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render(m.title))
	b.WriteString("\n")
	if m.subtitle != "" {
		b.WriteString(tuiHintStyle.Render(m.subtitle))
		b.WriteString("\n")
	}
	if m.allowReverse {
		order := "顺序"
		if m.reversed {
			order = "倒序"
		}
		b.WriteString(tuiMutedStyle.Render("当前排序: " + order))
		b.WriteString("\n")
	}
	if m.allowNumber {
		jump := "未输入"
		if m.numberBuffer != "" {
			jump = m.numberBuffer
		}
		b.WriteString(tuiMutedStyle.Render("直接输入序号: " + jump))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(m.items) == 0 {
		b.WriteString(tuiMutedStyle.Render(m.emptyMessage))
		return b.String()
	}

	start, end := m.visibleRange()
	for displayIndex := start; displayIndex < end; displayIndex++ {
		actual := m.displayIndexToActual(displayIndex)
		line := fmt.Sprintf("%s %d. %s", pointer(displayIndex == m.selected), displayIndex+1, m.items[actual])
		if displayIndex == m.selected {
			b.WriteString(tuiPickStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	if end < len(m.items) {
		b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("... 还有 %d 项", len(m.items)-end)))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m *selectorTUIModel) applyNumberBuffer() {
	if m.numberBuffer == "" {
		return
	}

	normalized := normalizeDigitInput(m.numberBuffer)
	for i, jumpValue := range m.jumpValues {
		if normalizeDigitInput(jumpValue) == normalized {
			m.selected = m.actualToDisplayIndex(i)
			return
		}
	}

	value, err := strconv.Atoi(normalized)
	if err == nil && value >= 1 && value <= len(m.items) {
		m.selected = value - 1
	}
}

func (m selectorTUIModel) displayIndexToActual(displayIndex int) int {
	if !m.reversed {
		return displayIndex
	}
	return len(m.items) - 1 - displayIndex
}

func (m selectorTUIModel) actualToDisplayIndex(actualIndex int) int {
	if !m.reversed {
		return actualIndex
	}
	return len(m.items) - 1 - actualIndex
}

func (m selectorTUIModel) visibleRange() (int, int) {
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

func (m selectorTUIModel) pageSize() int {
	if m.windowHeight <= 12 {
		return 10
	}
	return m.windowHeight - 8
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func normalizeDigitInput(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return strconv.FormatFloat(parsed, 'f', -1, 64)
	}
	return value
}
