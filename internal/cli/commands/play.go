package commands

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/ui"
)

func init() {
	Register(&PlayCommand{app: app.New()})
}

// PlayCommand 播放命令
type PlayCommand struct {
	app *app.App
}

// Name 返回命令名称
func (c *PlayCommand) Name() string {
	return "play"
}

// Run 执行命令
func (c *PlayCommand) Run(args []string) {
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

	// 选择动漫
	idx, err := ui.Select("选择动漫", len(results), func(i int) string { return results[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	// 获取剧集
	episodes, err := src.GetEpisodes(results[idx].URL)
	if err != nil {
		ui.Error("获取剧集失败: %v", err)
		os.Exit(1)
	}

	ui.Success("找到 %d 集", len(episodes))

	// 选择剧集
	epIdx, err := ui.Select("选择剧集", len(episodes), func(i int) string { return episodes[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	// 播放
	playEpisode(c.app, episodes[epIdx].URL)
}

// Usage 返回使用说明
func (c *PlayCommand) Usage() string {
	return "用法: goani play <keyword>\n示例: goani play 葬送的芙莉莲"
}

func playEpisode(application *app.App, episodeURL string) {
	src := application.GetFirstSource()
	videoURL, err := src.GetVideoURL(episodeURL)
	if err != nil {
		ui.Error("获取视频链接失败: %v", err)
		os.Exit(1)
	}
	playVideo(application, videoURL)
}
