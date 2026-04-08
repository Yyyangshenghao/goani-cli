package cli

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/cli/commands"

	// 导入 commands 包以触发 init() 自动注册
	_ "github.com/Yyyangshenghao/goani-cli/internal/cli/commands"
)

// Run CLI 入口
func Run() {
	// 获取所有已注册的命令
	cmds := commands.All()

	if len(os.Args) < 2 {
		printUsage(cmds)
		os.Exit(1)
	}

	cmdName := os.Args[1]
	args := os.Args[2:]

	// 支持简写
	if cmdName == "-v" || cmdName == "--version" {
		cmdName = "version"
	}

	cmd, exists := cmds[cmdName]
	if !exists {
		fmt.Printf("未知命令: %s\n", cmdName)
		printUsage(cmds)
		os.Exit(1)
	}

	cmd.Run(args)
}

func printUsage(cmds map[string]commands.Command) {
	fmt.Println("goani - 命令行动漫播放器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  goani <command> [arguments]")
	fmt.Println()
	fmt.Println("可用命令:")
	for name, cmd := range cmds {
		fmt.Printf("  %-10s %s\n", name, cmd.ShortDesc())
	}
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  goani search 葬送的芙莉莲")
	fmt.Println("  goani play 葬送的芙莉莲")
	fmt.Println("  goani config player mpv \"D:\\MPV播放器\\mpv.exe\"")
}
