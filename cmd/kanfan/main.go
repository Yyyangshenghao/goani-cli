package main

import (
    "fmt"
    "log"

    "github.com/yshscpu/goani-cli/internal/source"
)

func main() {
    config, err := source.LoadConfig("mediaSourceJson/css1.json")
    if err != nil {
        log.Fatal(err)
    }

    sources := config.ExportedMediaSourceDataList.MediaSources
    fmt.Printf("共加载 %d 个媒体源\n\n", len(sources))

    // 测试第一个源
    src := source.NewWebSelectorSource(sources[0])
    fmt.Printf("测试源: %s\n", src.Name())

    // 搜索
    results, err := src.Search("葬送的芙莉莲")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("搜索结果 (%d 条):\n", len(results))

    // 获取剧集
    if len(results) > 0 {
        episodes, err := src.GetEpisodes(results[2].URL)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("剧集列表 (%d 条)\n", len(episodes))

        // 获取视频直链
        if len(episodes) > 0 {
            fmt.Printf("\n获取视频直链: %s\n", episodes[0].Name)
            videoURL, err := src.GetVideoURL(episodes[0].URL)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Printf("视频直链: %s\n", videoURL)
        }
    }
}
