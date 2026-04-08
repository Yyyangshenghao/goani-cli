package commands

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/ui"
)

func init() {
	Register(&SearchCommand{app: app.New()})
}

// SearchCommand 搜索命令
type SearchCommand struct {
	app *app.App
}

// Name 返回命令名称
func (c *SearchCommand) Name() string {
	return "search"
}

// Run 执行命令
func (c *SearchCommand) Run(args []string) {
	if len(args) < 1 {
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	keyword := args[0]
	src := c.app.GetFirstSource()
	if src == nil {
		ui.Error("未找到媒体源")
		os.Exit(1)
	}

	ui.Info("搜索: %s", keyword)
	results, err := src.Search(keyword)
	if err != nil {
		ui.Error("搜索失败: %v", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		ui.Info("未找到结果")
		return
	}

	ui.Success("找到 %d 条结果", len(results))
	for i, r := range results {
		fmt.Printf("  %d. %s\n", i+1, r.Name)
	}

	// 选择动漫
	idx, err := ui.Select("选择动漫", len(results), func(i int) string { return results[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	showEpisodesAndSelect(c.app, results[idx].URL)
}

// Usage 返回使用说明
func (c *SearchCommand) Usage() string {
	return "用法: goani search <keyword>\n示例: goani search 葬送的芙莉莲"
}

// --- 辅助函数 ---

func showEpisodesAndSelect(application *app.App, animeURL string) {
	src := application.GetFirstSource()
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
		playVideo(application, videoURL)
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
