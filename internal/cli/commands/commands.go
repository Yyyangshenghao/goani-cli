package commands

import (
	"fmt"
	"os"

	"github.com/yshscpu/goani-cli/internal/app"
	"github.com/yshscpu/goani-cli/internal/source"
	"github.com/yshscpu/goani-cli/internal/ui"
)

// Search 搜索动漫
func Search(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: goani search <keyword>")
		os.Exit(1)
	}

	keyword := args[0]
	application := app.New()
	src := application.GetFirstSource()
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

	// 获取剧集
	showEpisodesAndSelect(application, results[idx].URL)
}

// Play 搜索并播放
func Play(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: goani play <keyword>")
		os.Exit(1)
	}

	keyword := args[0]
	application := app.New()
	src := application.GetFirstSource()
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

	// 获取剧集并播放
	episodes, err := src.GetEpisodes(results[idx].URL)
	if err != nil {
		ui.Error("获取剧集失败: %v", err)
		os.Exit(1)
	}

	ui.Success("找到 %d 集", len(episodes))

	epIdx, err := ui.Select("选择剧集", len(episodes), func(i int) string { return episodes[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	playEpisode(application, episodes[epIdx].URL)
}

// Config 配置播放器
func Config(args []string) {
	if len(args) < 3 || args[0] != "player" {
		fmt.Println("用法: goani config player <name> <path>")
		fmt.Println("示例: goani config player mpv \"D:\\MPV播放器\\mpv.exe\"")
		os.Exit(1)
	}

	name := args[1]
	path := args[2]

	application := app.New()
	application.Config.SetPlayer(name, path)
	if err := application.SaveConfig(); err != nil {
		ui.Error("保存配置失败: %v", err)
		os.Exit(1)
	}

	ui.Success("已配置播放器: %s", name)
}

// List 列出媒体源
func List() {
	application := app.New()
	ui.Info("共 %d 个媒体源", len(application.Sources))
	fmt.Println()
	for i, s := range application.Sources {
		fmt.Printf("  %d. %s\n", i+1, s.Arguments.Name)
	}
}

// showEpisodesAndSelect 显示剧集并选择
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

	// 获取视频直链
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

// playEpisode 播放剧集
func playEpisode(application *app.App, episodeURL string) {
	src := application.GetFirstSource()
	videoURL, err := src.GetVideoURL(episodeURL)
	if err != nil {
		ui.Error("获取视频链接失败: %v", err)
		os.Exit(1)
	}
	playVideo(application, videoURL)
}

// playVideo 播放视频
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

// printEpisodes 打印剧集列表
func printEpisodes(episodes []source.Episode, max int) {
	for i, ep := range episodes {
		if i >= max {
			fmt.Printf("  ... 还有 %d 集\n", len(episodes)-max)
			break
		}
		fmt.Printf("  %d. %s\n", i+1, ep.Name)
	}
}
