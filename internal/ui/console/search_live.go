package console

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
)

// SearchUI 是经典控制台模式下的实时搜索结果面板。
// 它不是完整 TUI，而是通过重绘终端文本来持续展示搜索进度。
// SearchUI 搜索界面状态
type SearchUI struct {
	keyword string
	total   int
	results []app.SourceSearchResult
	mu      sync.RWMutex
	updated chan struct{}
}

// NewSearchUI 创建搜索界面
func NewSearchUI(keyword string, totalSources int) *SearchUI {
	return &SearchUI{
		keyword: keyword,
		total:   totalSources,
		results: make([]app.SourceSearchResult, 0),
		updated: make(chan struct{}, 1),
	}
}

// AddResult 添加搜索结果并刷新显示
func (ui *SearchUI) AddResult(result app.SourceSearchResult) {
	ui.mu.Lock()
	ui.results = append(ui.results, result)
	ui.mu.Unlock()
	ui.notifyUpdated()
	ui.render()
}

// GetSuccessResults 获取所有成功的结果
func (ui *SearchUI) GetSuccessResults() []app.SourceSearchResult {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	var success []app.SourceSearchResult
	for _, r := range ui.results {
		if r.Status == app.StatusSuccess {
			success = append(success, r)
		}
	}
	return success
}

func (ui *SearchUI) snapshotResults() []app.SourceSearchResult {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	results := make([]app.SourceSearchResult, len(ui.results))
	copy(results, ui.results)
	return results
}

// render 渲染界面
func (ui *SearchUI) render() {
	results := ui.snapshotResults()
	success := successResults(results)

	// 清除当前显示
	ui.clear(results, success)

	// 打印标题
	fmt.Printf("搜索: %s\n", ui.keyword)

	// 打印每个源的状态
	for _, r := range results {
		ui.printSourceStatus(r)
	}

	// 打印等待中的源数量
	pending := ui.total - len(results)
	if pending > 0 {
		fmt.Printf("  ⏳ 等待中... (%d 个源)\n", pending)
	}

	// 打印可用源列表供选择
	if len(success) > 0 {
		fmt.Println()
		fmt.Println("选择片源 (输入序号，或等待更多结果):")
		for i, r := range success {
			fmt.Printf("  %d. %s    %dms  %d 条结果\n", i+1, r.SourceName, r.Duration.Milliseconds(), len(r.Results))
		}
		fmt.Println()
	}
}

// printSourceStatus 打印单个源的状态
func (ui *SearchUI) printSourceStatus(r app.SourceSearchResult) {
	switch r.Status {
	case app.StatusSuccess:
		fmt.Printf("  ✓ %s    %dms  找到 %d 条\n", r.SourceName, r.Duration.Milliseconds(), len(r.Results))
	case app.StatusTimeout:
		fmt.Printf("  ✗ %s    超时\n", r.SourceName)
	case app.StatusError:
		errMsg := "网络错误"
		if r.Error != nil {
			errMsg = truncateError(r.Error.Error())
		}
		fmt.Printf("  ✗ %s    %s\n", r.SourceName, errMsg)
	}
}

// clear 清除已显示的内容
func (ui *SearchUI) clear(results, success []app.SourceSearchResult) {
	lines := 2 + len(results) // 标题 + 结果行
	pending := ui.total - len(results)
	if pending > 0 {
		lines++ // 等待行
	}
	if len(success) > 0 {
		lines += 2 + len(success) + 1 // 空行 + 提示 + 选项 + 空行
	}

	// 移动光标到行首并清除
	for i := 0; i < lines; i++ {
		fmt.Print("\033[F\033[K") // 光标上移一行并清除该行
	}
}

// truncateError 截断错误信息
func truncateError(err string) string {
	if len(err) > 20 {
		return err[:20] + "..."
	}
	return err
}

// WaitForSelection 等待用户选择源
func (ui *SearchUI) WaitForSelection(done <-chan struct{}) int {
	for len(ui.GetSuccessResults()) == 0 {
		select {
		case <-ui.updated:
		case <-done:
			return -1
		}
	}

	// 读取用户输入
	inputChan := make(chan string, 1)
	go func() {
		var input string
		fmt.Fscan(os.Stdin, &input)
		inputChan <- strings.TrimSpace(input)
	}()

	select {
	case input := <-inputChan:
		success := ui.GetSuccessResults()
		idx := 0
		for _, c := range input {
			if c >= '0' && c <= '9' {
				idx = idx*10 + int(c-'0')
			}
		}
		if idx >= 1 && idx <= len(success) {
			return idx - 1
		}
		return -1
	case <-done:
		return -1
	}
}

func (ui *SearchUI) notifyUpdated() {
	select {
	case ui.updated <- struct{}{}:
	default:
	}
}

func successResults(results []app.SourceSearchResult) []app.SourceSearchResult {
	var success []app.SourceSearchResult
	for _, r := range results {
		if r.Status == app.StatusSuccess {
			success = append(success, r)
		}
	}
	return success
}
