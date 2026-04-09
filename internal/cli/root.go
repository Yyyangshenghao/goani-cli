package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

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
		return
	}

	cmdName := os.Args[1]
	args := os.Args[2:]

	if isHelpToken(cmdName) {
		if len(args) == 0 {
			printUsage(cmds)
			return
		}

		cmd, exists := cmds[args[0]]
		if !exists {
			fmt.Printf("未知命令: %s\n", args[0])
			printUsage(cmds)
			os.Exit(1)
		}

		fmt.Println(cmd.Usage())
		return
	}

	runCommand(cmds, cmdName, args)
}

func runCommand(cmds map[string]commands.Command, cmdName string, args []string) {
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

	for _, arg := range args {
		if isHelpToken(arg) {
			fmt.Println(cmd.Usage())
			return
		}
	}

	cmd.Run(args)
}

func printUsage(cmds map[string]commands.Command) {
	fmt.Println("goani - 命令行动漫播放器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  goani")
	fmt.Println("  goani <command> [arguments]")
	fmt.Println()
	fmt.Println("可用命令:")
	names := make([]string, 0, len(cmds))
	for name := range cmds {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cmd := cmds[name]
		fmt.Printf("  %-10s %s\n", name, cmd.ShortDesc())
	}
	fmt.Println()
	fmt.Println("推荐用法:")
	fmt.Println("  先看帮助和命令说明，再按需进入具体功能")
	fmt.Println("  想直接进入交互式界面时，使用 goani tui")
	fmt.Println("  想要实时交互界面时，使用 goani search --interactive")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  goani")
	fmt.Println("  goani tui")
	fmt.Println("  goani search 葬送的芙莉莲")
	fmt.Println("  goani search --interactive 葬送的芙莉莲")
	fmt.Println("  goani play 葬送的芙莉莲")
	fmt.Println("  goani config player mpv \"D:\\MPV播放器\\mpv.exe\"")
}

func isHelpToken(arg string) bool {
	switch strings.TrimSpace(arg) {
	case "help", "-h", "--help":
		return true
	default:
		return false
	}
}
