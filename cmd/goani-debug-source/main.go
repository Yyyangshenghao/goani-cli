package main

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

func main() {
	fmt.Println("=== goani 媒体源手动检查 ===")
	fmt.Println()

	fmt.Println("【1】加载媒体源")
	sm := webselector.NewTestSourceManager()
	sources := sm.GetAll()
	if len(sources) == 0 {
		fmt.Println("失败: 媒体源为空")
		os.Exit(1)
	}
	fmt.Printf("成功: 加载 %d 个媒体源\n\n", len(sources))

	fmt.Println("【2】搜索动漫")
	src := webselector.New(sources[0])
	fmt.Printf("测试源: %s\n", src.Name())
	results, err := src.Search("葬送的芙莉莲")
	if err != nil {
		fmt.Printf("失败: %v\n", err)
		os.Exit(1)
	}
	if len(results) == 0 {
		fmt.Println("失败: 搜索结果为空")
		os.Exit(1)
	}
	fmt.Printf("成功: 找到 %d 条结果\n", len(results))
	for i, r := range results {
		if i >= 3 {
			break
		}
		fmt.Printf("  - %s\n", r.Name)
	}
	fmt.Println()

	fmt.Println("【3】获取剧集列表")
	testAnime := results[0]
	fmt.Printf("测试动漫: %s\n", testAnime.Name)
	episodes, err := src.GetEpisodes(testAnime.URL)
	if err != nil {
		fmt.Printf("失败: %v\n", err)
		os.Exit(1)
	}
	if len(episodes) == 0 {
		fmt.Println("失败: 剧集列表为空")
		os.Exit(1)
	}
	fmt.Printf("成功: 找到 %d 集\n", len(episodes))
	for i, ep := range episodes {
		if i >= 5 {
			break
		}
		fmt.Printf("  - %s\n", ep.Name)
	}
	fmt.Println()

	fmt.Println("【4】获取视频直链")
	testEpisode := episodes[0]
	fmt.Printf("测试剧集: %s\n", testEpisode.Name)
	videoURL, err := src.GetVideoURL(testEpisode.URL)
	if err != nil {
		fmt.Printf("失败: %v\n", err)
		os.Exit(1)
	}
	if videoURL == "" {
		fmt.Println("失败: 视频直链为空")
		os.Exit(1)
	}
	fmt.Printf("成功: %s\n", videoURL)
}
