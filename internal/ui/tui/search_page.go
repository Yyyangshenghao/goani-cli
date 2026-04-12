package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
)

const searchDebounceDelay = 350 * time.Millisecond

var (
	tuiTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	tuiHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	tuiMutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	tuiErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	tuiOkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	tuiPickStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("62")).Bold(true)
)

// SearchTUISelection 表示搜索页完成后交给后续流程的选择结果。
// SearchTUISelection TUI 搜索完成后的选择结果
type SearchTUISelection struct {
	Results       []source.AggregatedAnime
	SelectedIndex int
}

type searchDebounceMsg struct {
	requestID int
	query     string
}

type searchResultMsg struct {
	requestID int
	ch        <-chan app.SourceSearchResult
	result    app.SourceSearchResult
}

// searchTUIModel 管理实时搜索页的输入、防抖请求和结果列表状态。
type searchTUIModel struct {
	app          *app.App
	input        textinput.Model
	spinner      spinner.Model
	totalSources int
	width        int
	height       int
	requestID    int
	activeSearch int
	completed    int
	results      []app.SourceSearchResult
	searching    bool
	selected     int
	selection    *SearchTUISelection
}

// SupportsInteractiveTUI 判断当前终端是否适合进入交互式 TUI
func SupportsInteractiveTUI() bool {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	return true
}

// RunSearchTUI 运行交互式实时搜索
func RunSearchTUI(application *app.App, initialKeyword string) (*SearchTUISelection, error) {
	model := newSearchTUIModel(application, initialKeyword)
	program := newProgram(model)
	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	result, ok := finalModel.(searchTUIModel)
	if !ok {
		return nil, fmt.Errorf("无法读取搜索界面状态")
	}

	return result.selection, nil
}

func newSearchTUIModel(application *app.App, initialKeyword string) searchTUIModel {
	input := textinput.New()
	input.Placeholder = "输入动漫名称，至少 2 个字"
	input.Focus()
	input.CharLimit = 64
	input.Width = 36
	input.SetValue(initialKeyword)

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

	return searchTUIModel{
		app:          application,
		input:        input,
		spinner:      spin,
		totalSources: application.SourceManager.EnabledCount(),
		selected:     0,
	}
}

func (m searchTUIModel) Init() tea.Cmd {
	if strings.TrimSpace(m.input.Value()) == "" {
		return textinput.Blink
	}
	m.requestID = 1
	return tea.Batch(textinput.Blink, debounceSearchCmd(m.requestID, m.input.Value()))
}

func (m searchTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.selection = nil
			return m, tea.Quit
		case "up", "k":
			aggregated := aggregatedSearchResults(m.results)
			if len(aggregated) > 0 {
				m.input.Blur()
			}
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "down", "j":
			aggregated := aggregatedSearchResults(m.results)
			if len(aggregated) > 0 {
				m.input.Blur()
			}
			if m.selected < len(aggregated)-1 {
				m.selected++
			}
			return m, nil
		case "enter":
			aggregated := aggregatedSearchResults(m.results)
			if len(aggregated) == 0 || m.selected < 0 || m.selected >= len(aggregated) {
				return m, nil
			}
			m.selection = &SearchTUISelection{
				Results:       aggregated,
				SelectedIndex: m.selected,
			}
			return m, tea.Quit
		}

		var cmds []tea.Cmd
		if !m.input.Focused() {
			cmds = append(cmds, m.input.Focus())
		}

		var cmd tea.Cmd
		before := m.input.Value()
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		after := m.input.Value()
		if before != after {
			m.requestID++
			m.activeSearch = 0
			m.selected = 0
			if strings.TrimSpace(after) == "" {
				m.results = nil
				m.completed = 0
				m.searching = false
				return m, tea.Batch(cmds...)
			}
			cmds = append(cmds, debounceSearchCmd(m.requestID, after))
			return m, tea.Batch(cmds...)
		}
		return m, tea.Batch(cmds...)

	case searchDebounceMsg:
		if msg.requestID != m.requestID {
			return m, nil
		}

		query := strings.TrimSpace(msg.query)
		if len([]rune(query)) < 2 {
			m.results = nil
			m.completed = 0
			m.selected = 0
			m.searching = false
			return m, nil
		}

		m.activeSearch = msg.requestID
		m.results = nil
		m.completed = 0
		m.selected = 0
		m.searching = true
		searchCh := m.app.SearchAll(query)
		return m, tea.Batch(m.spinner.Tick, waitSearchResultCmd(msg.requestID, searchCh))

	case searchResultMsg:
		if msg.requestID != m.activeSearch {
			return m, nil
		}

		m.results = append(m.results, msg.result)
		m.completed++
		aggregated := aggregatedSearchResults(m.results)
		if m.selected >= len(aggregated) && len(aggregated) > 0 {
			m.selected = len(aggregated) - 1
		}

		if m.completed >= m.totalSources {
			m.searching = false
			return m, nil
		}

		return m, waitSearchResultCmd(msg.requestID, msg.ch)

	case spinner.TickMsg:
		if !m.searching {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m searchTUIModel) View() string {
	var b strings.Builder

	title := "goani 实时搜索"
	if !SupportsInteractiveTUI() {
		title += " (已降级)"
	}
	b.WriteString(tuiTitleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(m.input.View())
	b.WriteString("\n")

	statusLine := m.renderStatusLine()
	if statusLine != "" {
		b.WriteString(statusLine)
		b.WriteString("\n")
	}

	aggregated := aggregatedSearchResults(m.results)
	if len(aggregated) == 0 {
		b.WriteString(tuiHintStyle.Render("输入至少 2 个字后开始搜索，按 Esc 退出"))
		if !m.searching && strings.TrimSpace(m.input.Value()) != "" && len([]rune(strings.TrimSpace(m.input.Value()))) >= 2 && m.completed >= m.totalSources {
			b.WriteString("\n")
			b.WriteString(tuiMutedStyle.Render("没有聚合到可用番剧，可以继续修改关键词"))
		}
		return b.String()
	}

	b.WriteString("\n")
	b.WriteString(tuiTitleStyle.Render("聚合番剧"))
	b.WriteString("\n")

	start, end := m.visibleResultRange(len(aggregated))
	if start > 0 {
		b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("... 上方还有 %d 项", start)))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		item := aggregated[i]
		line := fmt.Sprintf("%s  %s  [%d 个片源 / %d 条命中]", pointer(i == m.selected), item.Name, item.SourceCount(), item.HitCount())
		if i == m.selected {
			b.WriteString(tuiPickStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	if end < len(aggregated) {
		b.WriteString(tuiMutedStyle.Render(fmt.Sprintf("... 下方还有 %d 项", len(aggregated)-end)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(tuiHintStyle.Render("↑/↓ 选择番剧，Enter 确认，Esc 退出"))

	return strings.TrimRight(b.String(), "\n")
}

func (m searchTUIModel) renderStatusLine() string {
	query := strings.TrimSpace(m.input.Value())
	if query == "" {
		return tuiMutedStyle.Render("等待输入关键词")
	}
	if len([]rune(query)) < 2 {
		return tuiMutedStyle.Render("请至少输入 2 个字")
	}
	if m.searching {
		return fmt.Sprintf("%s 正在搜索 %q  (%d/%d)", m.spinner.View(), query, m.completed, m.totalSources)
	}
	return tuiMutedStyle.Render(fmt.Sprintf("搜索完成 %q  (%d/%d)", query, m.completed, m.totalSources))
}

func debounceSearchCmd(requestID int, query string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(searchDebounceDelay)
		return searchDebounceMsg{
			requestID: requestID,
			query:     query,
		}
	}
}

func waitSearchResultCmd(requestID int, ch <-chan app.SourceSearchResult) tea.Cmd {
	return func() tea.Msg {
		result := <-ch
		return searchResultMsg{
			requestID: requestID,
			ch:        ch,
			result:    result,
		}
	}
}

func successSearchResults(results []app.SourceSearchResult) []app.SourceSearchResult {
	var success []app.SourceSearchResult
	for _, item := range results {
		if item.Status == app.StatusSuccess && len(item.Results) > 0 {
			success = append(success, item)
		}
	}
	sort.SliceStable(success, func(i, j int) bool {
		if success[i].SourcePriority != success[j].SourcePriority {
			return success[i].SourcePriority > success[j].SourcePriority
		}
		if success[i].Duration != success[j].Duration {
			return success[i].Duration < success[j].Duration
		}
		return success[i].SourceName < success[j].SourceName
	})
	return success
}

func aggregatedSearchResults(results []app.SourceSearchResult) []source.AggregatedAnime {
	success := successSearchResults(results)
	hits := make([]source.AnimeHit, 0)
	for _, item := range success {
		for _, anime := range item.Results {
			hits = append(hits, source.AnimeHit{
				SourceName: item.SourceName,
				Anime:      anime,
			})
		}
	}
	return source.GroupAnimes(hits)
}

func (m searchTUIModel) visibleResultRange(total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}

	page := m.resultPageSize()
	start := m.selected - page/2
	if start < 0 {
		start = 0
	}

	end := start + page
	if end > total {
		end = total
		start = end - page
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

func (m searchTUIModel) resultPageSize() int {
	if m.height <= 0 {
		return 5
	}

	page := m.height - 9
	if page < 1 {
		page = 1
	}
	if page > 10 {
		page = 10
	}
	return page
}

func pointer(selected bool) string {
	if selected {
		return ">"
	}
	return " "
}
