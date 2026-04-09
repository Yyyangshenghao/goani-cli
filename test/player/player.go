package main

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
)

func main() {
	fmt.Println("=== 播放器模块测试 ===")
	fmt.Println()

	// 测试 1: 加载配置
	fmt.Println("【测试 1】加载配置")
	cfg, err := player.LoadConfig()
	if err != nil {
		fmt.Printf("❌ 失败: %v\n\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 成功\n")
	if cfg.DefaultPlayer != "" {
		fmt.Printf("   默认播放器: %s\n", cfg.DefaultPlayer)
		if path := cfg.GetPath(cfg.DefaultPlayer); path != "" {
			fmt.Printf("   路径: %s\n", path)
		}
	} else {
		fmt.Println("   未配置播放器（将自动检测）")
	}
	fmt.Println()

	// 测试 2: 自动检测播放器
	fmt.Println("【测试 2】自动检测播放器")
	pm := player.NewManager()
	available := pm.GetAvailable()
	if len(available) == 0 {
		fmt.Println("⚠️  警告: 未检测到可用播放器")
		fmt.Println("   请运行: goani config player <name> <path>")
	} else {
		fmt.Printf("✅ 成功: 检测到 %d 个播放器\n", len(available))
		for _, p := range available {
			fmt.Printf("   - %s\n", p.Name())
		}
	}
	fmt.Println()

	// 测试 3: 使用配置的播放器
	fmt.Println("【测试 3】使用配置的播放器")
	pmWithConfig := player.NewManagerWithConfig(cfg.DefaultPlayer, cfg.Paths)
	configPlayer := pmWithConfig.GetFirst()
	if configPlayer == nil {
		fmt.Println("❌ 失败: 无可用播放器")
		fmt.Println()
		os.Exit(1)
	}
	fmt.Printf("✅ 成功: %s\n\n", configPlayer.Name())

	fmt.Println("=== 所有测试通过 ===")
}
