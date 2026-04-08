package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yshscpu/goani-cli/internal/config"
	"github.com/yshscpu/goani-cli/internal/player"
	"github.com/yshscpu/goani-cli/internal/source"
	"github.com/yshscpu/goani-cli/internal/source/webselector"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 检查播放器
	pm := player.NewManagerWithConfig(cfg.PlayerName, cfg.PlayerPath)
	available := pm.GetAvailable()

	if len(available) == 0 {
		fmt.Println("未找到可用的播放器")
		fmt.Println("请运行以下命令配置播放器:")
		fmt.Println("  goani config player mpv \"D:\\path\\to\\mpv.exe\"")
		os.Exit(1)
	}

	fmt.Printf("可用播放器: ")
	for _, p := range available {
		fmt.Printf("%s ", p.Name())
	}
	fmt.Println("\n")

	// 加载媒体源
	srcConfig, err := source.LoadConfig("mediaSourceJson/css1.json")
	if err != nil {
		log.Fatal(err)
	}

	sources := srcConfig.ExportedMediaSourceDataList.MediaSources
	fmt.Printf("共加载 %d 个媒体源\n\n", len(sources))

	// 测试第一个源
	src := webselector.New(sources[0])
	fmt.Printf("测试源: %s\n", src.Name())

	// 搜索
	results, err := src.Search("葬送的芙莉莲")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("搜索结果 (%d 条)\n", len(results))

	// 获取剧集
	if len(results) > 0 {
		episodes, err := src.GetEpisodes(results[2].URL)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("剧集列表 (%d 条)\n", len(episodes))

		// 获取视频直链并播放
		if len(episodes) > 0 {
			fmt.Printf("\n获取视频直链: %s\n", episodes[0].Name)
			videoURL, err := src.GetVideoURL(episodes[0].URL)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("视频直链: %s\n", videoURL)

			// 播放视频
			p := pm.GetFirst()
			fmt.Printf("\n使用 %s 播放...\n", p.Name())
			if err := p.Play(videoURL); err != nil {
				log.Fatal(err)
			}
			fmt.Println("播放器已启动")
		}
	}
}
