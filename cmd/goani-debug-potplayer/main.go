package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
)

const defaultM3U8URL = "https://vip.dytt-hot.com/20250211/59942_6939555e/index.m3u8"

func main() {
	url := defaultM3U8URL
	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) != "" {
		url = strings.TrimSpace(os.Args[1])
	}

	cfg, err := player.LoadConfig()
	if err != nil {
		fmt.Printf("加载播放器配置失败: %v\n", err)
		os.Exit(1)
	}

	potPath := cfg.GetPath("potplayer")
	if strings.TrimSpace(potPath) == "" {
		fmt.Println("未配置 PotPlayer 路径")
		fmt.Println("请先运行: goani config player potplayer <path>")
		os.Exit(1)
	}

	p := player.NewManagerWithConfig("potplayer", cfg.Paths).GetByName("potplayer")
	if p == nil {
		fmt.Println("当前未检测到可用的 PotPlayer")
		fmt.Printf("配置路径: %s\n", potPath)
		os.Exit(1)
	}

	fmt.Println("=== PotPlayer 直连 m3u8 手动检查 ===")
	fmt.Printf("PotPlayer: %s\n", potPath)
	fmt.Printf("URL: %s\n", url)
	fmt.Println()
	fmt.Println("正在直接调用 PotPlayer 打开该链接...")

	if err := p.Play(url); err != nil {
		fmt.Printf("启动失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("已发起播放。")
	fmt.Println("如果这里仍然报错，而 goani 正常播放，说明差异主要在 goani 的本地 HLS 代理兼容层。")
}
