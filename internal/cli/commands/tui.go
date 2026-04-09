package commands

import (
	"os"
	"strings"

	consoleui "github.com/Yyyangshenghao/goani-cli/internal/ui/console"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
	"github.com/Yyyangshenghao/goani-cli/internal/version"
	"github.com/Yyyangshenghao/goani-cli/internal/workflow"
)

func init() {
	Register(&TUICommand{})
}

// TUICommand 交互式 TUI 入口
type TUICommand struct{}

// Name 返回命令名称
func (c *TUICommand) Name() string {
	return "tui"
}

// ShortDesc 返回简短描述
func (c *TUICommand) ShortDesc() string {
	return "进入交互式 TUI 模式"
}

// Run 执行命令
func (c *TUICommand) Run(args []string) {
	if !tui.SupportsInteractiveTUI() {
		consoleui.Error("当前终端不支持 TUI，请改用 goani search 或 goani play")
		os.Exit(1)
	}

	keyword := strings.TrimSpace(strings.Join(args, " "))
	searchCmd := &SearchCommand{}
	application := searchCmd.ensureApp()

	for {
		action, err := tui.RunMainTUI()
		if err != nil {
			consoleui.Error("启动 TUI 失败: %v", err)
			os.Exit(1)
		}

		switch action {
		case tui.MainMenuActionSearch:
			if err := searchCmd.runInteractiveSearchWithError(application, keyword); err != nil {
				if pageErr := tui.RunTextTUI("搜索失败", err.Error()); pageErr != nil {
					consoleui.Error("搜索失败: %v", err)
					os.Exit(1)
				}
			}
			keyword = ""
		case tui.MainMenuActionSourceList:
			if err := tui.RunTextTUI("媒体源", workflow.RenderSourceOverview(application)); err != nil {
				consoleui.Error("打开媒体源页面失败: %v", err)
				os.Exit(1)
			}
		case tui.MainMenuActionConfig:
			if err := workflow.RunConfigTUI(application); err != nil {
				consoleui.Error("打开配置页面失败: %v", err)
				os.Exit(1)
			}
		case tui.MainMenuActionVersion:
			if err := tui.RunTextTUI("版本信息", version.Info()); err != nil {
				consoleui.Error("打开版本信息失败: %v", err)
				os.Exit(1)
			}
		case tui.MainMenuActionQuit:
			return
		default:
			return
		}
	}
}

// Usage 返回使用说明
func (c *TUICommand) Usage() string {
	return "用法: goani tui [keyword]\n\n进入交互式 TUI 模式。\n首页可查看搜索、媒体源、配置和版本信息。\n如果提供关键词，会在进入搜索页时自动带入。\n\n示例:\n  goani tui\n  goani tui 葬送的芙莉莲"
}
