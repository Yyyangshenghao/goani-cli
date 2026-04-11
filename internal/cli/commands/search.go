package commands

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	consoleui "github.com/Yyyangshenghao/goani-cli/internal/ui/console"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
	"github.com/Yyyangshenghao/goani-cli/internal/workflow"
)

func init() {
	Register(&SearchCommand{})
}

// SearchCommand 搜索命令
type SearchCommand struct {
	app *app.App
}

func (c *SearchCommand) ensureApp() *app.App {
	if c.app == nil {
		c.app = app.New()
	}
	return c.app
}

// Name 返回命令名称
func (c *SearchCommand) Name() string {
	return "search"
}

// ShortDesc 返回简短描述
func (c *SearchCommand) ShortDesc() string {
	return "搜索动漫（支持实时 TUI）"
}

// Run 执行命令
func (c *SearchCommand) Run(args []string) {
	interactive, keyword := parseSearchArgs(args)
	if keyword == "" && !interactive {
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	application := c.ensureApp()
	if interactive {
		c.runInteractiveSearch(application, keyword)
		return
	}

	c.runClassicSearch(application, keyword)
}

func (c *SearchCommand) runClassicSearch(application *app.App, keyword string) {
	totalSources := application.SourceManager.Count()
	if totalSources == 0 {
		consoleui.Error("未找到媒体源")
		os.Exit(1)
	}

	// 创建搜索界面
	searchUI := consoleui.NewSearchUI(keyword, totalSources)

	// 启动并发搜索
	resultChan := application.SearchAll(keyword)

	// 用于通知用户输入监听器退出
	var wg sync.WaitGroup
	done := make(chan struct{})
	selectedChan := make(chan int, 1)

	// 启动用户输入监听
	wg.Add(1)
	go func() {
		defer wg.Done()
		if idx := searchUI.WaitForSelection(done); idx >= 0 {
			selectedChan <- idx
		}
	}()

	// 收集结果
	var completed int
	var selectedIndex int = -1

loop:
	for completed < totalSources {
		select {
		case result := <-resultChan:
			completed++
			searchUI.AddResult(result)

			// 检查是否所有源都失败了
			if completed == totalSources {
				success := searchUI.GetSuccessResults()
				if len(success) == 0 {
					close(done)
					wg.Wait()
					consoleui.Error("所有源都搜索失败，请检查网络或稍后重试")
					os.Exit(1)
				}
			}

		case idx := <-selectedChan:
			selectedIndex = idx
			close(done)
			break loop
		}
	}

	wg.Wait()

	// 如果用户没有选择，让用户从成功列表中选择
	success := searchUI.GetSuccessResults()
	if selectedIndex < 0 {
		if len(success) == 0 {
			consoleui.Error("没有可用的搜索结果")
			os.Exit(1)
		}
		selectedIndex, _ = consoleui.Select("选择片源查看结果", len(success), func(i int) string {
			r := success[i]
			return fmt.Sprintf("%s    %dms  %d 条结果", r.SourceName, r.Duration.Milliseconds(), len(r.Results))
		})
	}

	if selectedIndex < 0 || selectedIndex >= len(success) {
		fmt.Println("已取消")
		return
	}

	// 显示选中源的结果
	selectedResult := success[selectedIndex]
	consoleui.Success("已选择: %s", selectedResult.SourceName)
	workflow.ShowAnimeListAndSelect(application, selectedResult.Results, selectedResult.SourceName)
}

func (c *SearchCommand) runInteractiveSearch(application *app.App, keyword string) {
	if err := c.runInteractiveSearchWithError(application, keyword); err != nil {
		consoleui.Error("%v", err)
		os.Exit(1)
	}
}

// runInteractiveSearchWithError 保留在命令层，负责 TUI 能力判断和 classic fallback；
// 真正的交互流程已经下沉到 workflow 包。
func (c *SearchCommand) runInteractiveSearchWithError(application *app.App, keyword string) error {
	if application.SourceManager.Count() == 0 {
		return fmt.Errorf("未找到媒体源")
	}

	if !tui.SupportsInteractiveTUI() {
		if keyword == "" {
			keyword = consoleui.Input("输入关键词: ")
		}
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			fmt.Println("已取消")
			return nil
		}
		consoleui.Info("当前终端不支持交互式 TUI，已切换到普通搜索模式")
		c.runClassicSearch(application, keyword)
		return nil
	}

	selection, err := tui.RunSearchTUI(application, keyword)
	if err != nil {
		return fmt.Errorf("启动实时搜索失败: %w", err)
	}
	if selection == nil {
		return nil
	}

	return workflow.ShowInteractiveSelectionFlow(application, selection.Results, selection.SelectedIndex)
}

// Usage 返回使用说明
func (c *SearchCommand) Usage() string {
	return "用法: goani search [--interactive|-i] <keyword>\n\n参数:\n  --interactive, -i    启用实时交互式搜索界面（TUI）\n\n示例:\n  goani search 葬送的芙莉莲\n  goani search --interactive 葬送的芙莉莲\n  goani search -i 葬送的芙莉莲"
}

func parseSearchArgs(args []string) (bool, string) {
	interactive := false
	keywordParts := make([]string, 0, len(args))

	for _, arg := range args {
		switch arg {
		case "--interactive", "-i":
			interactive = true
		default:
			keywordParts = append(keywordParts, arg)
		}
	}

	return interactive, strings.TrimSpace(strings.Join(keywordParts, " "))
}
