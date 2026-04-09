package main

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
)

func main() {
	fmt.Println("=== goani 播放器手动检查 ===")
	fmt.Println()

	fmt.Println("【1】加载播放器配置")
	cfg, err := player.LoadConfig()
	if err != nil {
		fmt.Printf("失败: %v\n\n", err)
		os.Exit(1)
	}
	fmt.Println("成功")
	if cfg.DefaultPlayer != "" {
		fmt.Printf("默认播放器: %s\n", cfg.DefaultPlayer)
		if path := cfg.GetPath(cfg.DefaultPlayer); path != "" {
			fmt.Printf("路径: %s\n", path)
		}
	} else {
		fmt.Println("未配置默认播放器（运行时会尝试自动挑一个已配置播放器）")
	}
	fmt.Println()

	fmt.Println("【2】检测当前可用播放器")
	pm := player.NewManager()
	available := pm.GetAvailable()
	if len(available) == 0 {
		fmt.Println("未检测到可用播放器")
		fmt.Println("请先运行: goani config player <name> <path>")
	} else {
		fmt.Printf("检测到 %d 个播放器:\n", len(available))
		for _, p := range available {
			fmt.Printf("  - %s\n", p.Name())
		}
	}
	fmt.Println()

	fmt.Println("【3】按当前配置挑选播放器")
	pmWithConfig := player.NewManagerWithConfig(cfg.DefaultPlayer, cfg.Paths)
	configPlayer := pmWithConfig.GetFirst()
	if configPlayer == nil {
		fmt.Println("失败: 当前没有可用播放器")
		os.Exit(1)
	}
	fmt.Printf("最终会使用: %s\n", configPlayer.Name())
}
