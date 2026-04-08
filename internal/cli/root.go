package cli

import (
	"fmt"
	"os"

	"github.com/yshscpu/goani-cli/internal/cli/commands"
)

var version = "0.1.0"

// Run CLI 入口
func Run() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "search":
		commands.Search(args)
	case "play":
		commands.Play(args)
	case "config":
		commands.Config(args)
	case "list":
		commands.List()
	case "version", "-v", "--version":
		fmt.Printf("goani v%s\n", version)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("goani - 命令行动漫播放器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  goani search <keyword>             搜索动漫")
	fmt.Println("  goani play <keyword>               搜索并播放")
	fmt.Println("  goani config player <name> <path>  配置播放器")
	fmt.Println("  goani list                         列出媒体源")
	fmt.Println("  goani version                      显示版本")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  goani search 葬送的芙莉莲")
	fmt.Println("  goani play 葬送的芙莉莲")
	fmt.Println("  goani config player mpv \"D:\\MPV播放器\\mpv.exe\"")
}
