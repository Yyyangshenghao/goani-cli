package commands

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
	"github.com/Yyyangshenghao/goani-cli/internal/ui"
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
	return "搜索动漫"
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
		ui.Error("未找到媒体源")
		os.Exit(1)
	}

	// 创建搜索界面
	searchUI := ui.NewSearchUI(keyword, totalSources)

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
					ui.Error("所有源都搜索失败，请检查网络或稍后重试")
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
			ui.Error("没有可用的搜索结果")
			os.Exit(1)
		}
		selectedIndex, _ = ui.Select("选择片源查看结果", len(success), func(i int) string {
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
	ui.Success("已选择: %s", selectedResult.SourceName)
	showAnimeListAndSelect(application, selectedResult.Results, selectedResult.SourceName)
}

func (c *SearchCommand) runInteractiveSearch(application *app.App, keyword string) {
	if application.SourceManager.Count() == 0 {
		ui.Error("未找到媒体源")
		os.Exit(1)
	}

	if !ui.SupportsInteractiveTUI() {
		if keyword == "" {
			keyword = ui.Input("输入关键词: ")
		}
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			fmt.Println("已取消")
			return
		}
		ui.Info("当前终端不支持交互式 TUI，已切换到普通搜索模式")
		c.runClassicSearch(application, keyword)
		return
	}

	selection, err := ui.RunSearchTUI(application, keyword)
	if err != nil {
		ui.Error("启动实时搜索失败: %v", err)
		os.Exit(1)
	}
	if selection == nil {
		fmt.Println("已取消")
		return
	}

	ui.Success("已选择: %s", selection.SourceName)
	showAnimeListAndSelect(application, selection.Results, selection.SourceName)
}

// Usage 返回使用说明
func (c *SearchCommand) Usage() string {
	return "用法: goani search [--interactive|-i] <keyword>\n示例: goani search 葬送的芙莉莲\n      goani search --interactive 葬送的芙莉莲"
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

// --- 辅助函数 ---

func showAnimeListAndSelect(application *app.App, animes []source.Anime, sourceName string) {
	if len(animes) == 0 {
		ui.Info("未找到结果")
		return
	}

	ui.Success("[%s] 找到 %d 条结果", sourceName, len(animes))
	for i, a := range animes {
		fmt.Printf("  %d. %s\n", i+1, a.Name)
	}

	idx, err := ui.Select("选择动漫", len(animes), func(i int) string { return animes[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	// 获取选中动漫的源实例
	src := application.GetSourceByName(sourceName)
	if src == nil {
		ui.Error("无法获取媒体源")
		os.Exit(1)
	}

	showEpisodesAndSelectWithSource(src, animes[idx].URL)
}

func showEpisodesAndSelectWithSource(src *webselector.WebSelectorSource, animeURL string) {
	episodes, err := src.GetEpisodes(animeURL)
	if err != nil {
		ui.Error("获取剧集失败: %v", err)
		os.Exit(1)
	}

	ui.Success("找到 %d 集", len(episodes))
	printEpisodes(episodes, 20)

	epIdx, err := ui.Select("选择剧集", len(episodes), func(i int) string { return episodes[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	videoURL, err := src.GetVideoURL(episodes[epIdx].URL)
	if err != nil {
		ui.Error("获取视频链接失败: %v", err)
		os.Exit(1)
	}

	ui.Success("视频链接: %s", videoURL)

	if ui.Confirm("是否播放?") {
		playVideoWithApp(videoURL)
	}
}

func printEpisodes(episodes []source.Episode, max int) {
	for i, ep := range episodes {
		if i >= max {
			fmt.Printf("  ... 还有 %d 集\n", len(episodes)-max)
			break
		}
		fmt.Printf("  %d. %s\n", i+1, ep.Name)
	}
}

func playVideoWithApp(url string) {
	application := app.New()
	p := application.GetPlayer()
	if p == nil {
		ui.Error("未找到可用播放器")
		fmt.Println("请先配置: goani config player <name> <path>")
		os.Exit(1)
	}

	ui.Info("使用 %s 播放...", p.Name())
	if err := p.Play(url); err != nil {
		ui.Error("播放失败: %v", err)
		os.Exit(1)
	}
	ui.Success("播放器已启动")
}

// playVideo 兼容 play.go 的调用
func playVideo(application *app.App, url string) {
	p := application.GetPlayer()
	if p == nil {
		ui.Error("未找到可用播放器")
		fmt.Println("请先配置: goani config player <name> <path>")
		os.Exit(1)
	}

	ui.Info("使用 %s 播放...", p.Name())
	if err := p.Play(url); err != nil {
		ui.Error("播放失败: %v", err)
		os.Exit(1)
	}
	ui.Success("播放器已启动")
}
