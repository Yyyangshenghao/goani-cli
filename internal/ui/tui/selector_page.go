package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
)

type selectorTUIResult struct {
	index int
}

// selectorTUIModel 是通用列表选择页，服务于番剧选择和选集选择两个场景。
type selectorTUIModel struct {
	title        string
	subtitle     string
	emptyMessage string
	items        []string
	jumpValues   []string
	showOrdinal  bool
	allowFilter  bool
	allowReverse bool
	allowNumber  bool
	reversed     bool
	selected     int
	numberBuffer string
	filterInput  textinput.Model
	filtered     []int
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

	filterInput := textinput.New()
	filterInput.Placeholder = "在当前结果里继续搜索"
	filterInput.CharLimit = 64
	filterInput.Width = 36
	filterInput.Focus()

	model := selectorTUIModel{
		title:        fmt.Sprintf("片源: %s", sourceName),
		subtitle:     "输入关键字过滤当前结果，Enter 确认，Esc 返回片源列表",
		emptyMessage: "当前片源没有可选番剧",
		items:        items,
		jumpValues:   jumpValues,
		showOrdinal:  true,
		allowFilter:  true,
		filterInput:  filterInput,
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, err
	}

	return &animes[result.index], nil
}

// RunAggregatedAnimeSelectionTUI 运行跨片源聚合后的番剧选择页。
func RunAggregatedAnimeSelectionTUI(animes []source.AggregatedAnime, initialIndex int) (*source.AggregatedAnime, int, error) {
	items := make([]string, len(animes))
	jumpValues := make([]string, len(animes))
	for i, anime := range animes {
		items[i] = aggregatedAnimeSelectionLabel(anime)
		jumpValues[i] = strconv.Itoa(i + 1)
	}

	filterInput := textinput.New()
	filterInput.Placeholder = "在聚合结果里继续搜索"
	filterInput.CharLimit = 64
	filterInput.Width = 36
	filterInput.Focus()

	model := selectorTUIModel{
		title:        "聚合番剧",
		subtitle:     "输入关键字过滤当前结果，Enter 确认，Esc 返回上一层",
		emptyMessage: "当前搜索结果没有可选番剧",
		items:        items,
		jumpValues:   jumpValues,
		showOrdinal:  true,
		allowFilter:  true,
		filterInput:  filterInput,
		selected:     clampSelectionIndex(initialIndex, len(animes)),
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, -1, err
	}

	return &animes[result.index], result.index, nil
}

// RunEpisodeSelectionTUI 运行剧集选择 TUI
func RunEpisodeSelectionTUI(animeName string, episodes []source.EpisodeGroup) (*source.EpisodeGroup, error) {
	result, _, err := RunEpisodeSelectionTUIWithSelection(animeName, episodes, 0)
	return result, err
}

// RunEpisodeSelectionTUIWithSelection 运行剧集选择 TUI，并允许调用方指定返回后默认高亮的剧集。
func RunEpisodeSelectionTUIWithSelection(animeName string, episodes []source.EpisodeGroup, initialIndex int) (*source.EpisodeGroup, int, error) {
	items := make([]string, len(episodes))
	jumpValues := make([]string, len(episodes))
	for i, episode := range episodes {
		items[i] = episodeSelectionLabel(episode)
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
		showOrdinal:  false,
		allowReverse: true,
		allowNumber:  true,
		reversed:     false,
		selected:     clampSelectionIndex(initialIndex, len(episodes)),
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, -1, err
	}

	return &episodes[result.index], result.index, nil
}

// runSelectorTUI 统一启动列表类 TUI，避免番剧和剧集各自维护一套状态机。
func runSelectorTUI(model selectorTUIModel) (*selectorTUIResult, error) {
	model.applyFilter()
	program := newProgram(model)
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
			if m.selected < m.filteredLen()-1 {
				m.selected++
			}
			return m, nil
		case "home":
			m.selected = 0
			return m, nil
		case "end":
			if m.filteredLen() > 0 {
				m.selected = m.filteredLen() - 1
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
			if m.allowFilter {
				before := m.filterInput.Value()
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				if before != m.filterInput.Value() {
					m.applyFilter()
				}
				return m, cmd
			}
			if m.allowNumber && m.numberBuffer != "" {
				m.numberBuffer = m.numberBuffer[:len(m.numberBuffer)-1]
				m.applyNumberBuffer()
			}
			return m, nil
		case "enter":
			if m.filteredLen() == 0 {
				return m, nil
			}
			if m.allowNumber && m.numberBuffer != "" {
				m.applyNumberBuffer()
			}
			m.result = &selectorTUIResult{index: m.displayIndexToActual(m.selected)}
			return m, tea.Quit
		}

		if m.allowFilter {
			before := m.filterInput.Value()
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			if before != m.filterInput.Value() {
				m.applyFilter()
				return m, cmd
			}
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
	if m.allowFilter {
		b.WriteString(m.filterInput.View())
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

	if m.filteredLen() == 0 {
		b.WriteString(tuiMutedStyle.Render(m.emptyMessage))
		if m.allowFilter && strings.TrimSpace(m.filterInput.Value()) != "" {
			b.WriteString("\n")
			b.WriteString(tuiMutedStyle.Render("没有匹配项，继续修改过滤关键词试试"))
		}
		return b.String()
	}

	start, end := m.visibleRange()
	for displayIndex := start; displayIndex < end; displayIndex++ {
		actual := m.displayIndexToActual(displayIndex)
		line := fmt.Sprintf("%s %s", pointer(displayIndex == m.selected), m.items[actual])
		if m.showOrdinal {
			line = fmt.Sprintf("%s %d. %s", pointer(displayIndex == m.selected), displayIndex+1, m.items[actual])
		}
		if displayIndex == m.selected {
			b.WriteString(tuiPickStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	if end < m.filteredLen() {
		b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("... 还有 %d 项", m.filteredLen()-end)))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// applyNumberBuffer 支持用户直接输入集数/序号跳转，并兼容倒序显示。
func (m *selectorTUIModel) applyNumberBuffer() {
	if m.numberBuffer == "" {
		return
	}

	normalized := normalizeDigitInput(m.numberBuffer)
	for i, jumpValue := range m.jumpValues {
		if !m.containsActualIndex(i) {
			continue
		}
		if normalizeDigitInput(jumpValue) == normalized {
			m.selected = m.actualToDisplayIndex(i)
			return
		}
	}

	value, err := strconv.Atoi(normalized)
	if err == nil && value >= 1 && value <= m.filteredLen() {
		m.selected = value - 1
	}
}

// displayIndexToActual / actualToDisplayIndex 负责在顺序和倒序视图之间映射真实索引。
func (m selectorTUIModel) displayIndexToActual(displayIndex int) int {
	if len(m.filtered) == 0 {
		return displayIndex
	}
	if !m.reversed {
		return m.filtered[displayIndex]
	}
	return m.filtered[len(m.filtered)-1-displayIndex]
}

func (m selectorTUIModel) actualToDisplayIndex(actualIndex int) int {
	for displayIndex := 0; displayIndex < len(m.filtered); displayIndex++ {
		if m.displayIndexToActual(displayIndex) == actualIndex {
			return displayIndex
		}
	}
	return 0
}

// visibleRange 只渲染当前视窗附近的内容，避免长列表把终端刷满。
func (m selectorTUIModel) visibleRange() (int, int) {
	page := m.pageSize()
	start := m.selected - page/2
	if start < 0 {
		start = 0
	}
	end := start + page
	if end > m.filteredLen() {
		end = m.filteredLen()
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

// isDigits 只允许集数跳转缓存接收数字键，避免和普通快捷键混在一起。
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

// normalizeDigitInput 让 01 / 1 / 1.0 这类输入跳到同一集。
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

func (m *selectorTUIModel) applyFilter() {
	m.filtered = m.filtered[:0]

	query := strings.ToLower(strings.TrimSpace(m.filterInput.Value()))
	for i, item := range m.items {
		if query == "" || strings.Contains(strings.ToLower(item), query) {
			m.filtered = append(m.filtered, i)
		}
	}

	if len(m.filtered) == 0 {
		m.selected = 0
		return
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func (m selectorTUIModel) filteredLen() int {
	if len(m.filtered) == 0 {
		return 0
	}
	return len(m.filtered)
}

func (m selectorTUIModel) containsActualIndex(actualIndex int) bool {
	for _, idx := range m.filtered {
		if idx == actualIndex {
			return true
		}
	}
	return false
}

// episodeSelectionLabel 让选集页优先展示真实集号，而不是再额外叠加一层列表编号。
func episodeSelectionLabel(episode source.EpisodeGroup) string {
	label := strings.TrimSpace(episode.Name)
	if episode.HasNumber && strings.TrimSpace(episode.Number) != "" {
		label = normalizeDigitInput(episode.Number)
	}
	if len(episode.Candidates) > 1 {
		return fmt.Sprintf("%s  (%d 条线路)", label, len(episode.Candidates))
	}
	return label
}

func aggregatedAnimeSelectionLabel(anime source.AggregatedAnime) string {
	return fmt.Sprintf("%s  (%d 个片源 / %d 条命中)", anime.Name, anime.SourceCount(), anime.HitCount())
}

func clampSelectionIndex(index, total int) int {
	if total <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= total {
		return total - 1
	}
	return index
}
