package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/version"
	"github.com/Yyyangshenghao/goani-cli/internal/ui"
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
	if !ui.SupportsInteractiveTUI() {
		ui.Error("当前终端不支持 TUI，请改用 goani search 或 goani play")
		os.Exit(1)
	}

	keyword := strings.TrimSpace(strings.Join(args, " "))
	searchCmd := &SearchCommand{}
	application := searchCmd.ensureApp()

	for {
		action, err := ui.RunMainTUI()
		if err != nil {
			ui.Error("启动 TUI 失败: %v", err)
			os.Exit(1)
		}

		switch action {
		case ui.MainMenuActionSearch:
			if err := searchCmd.runInteractiveSearchWithError(application, keyword); err != nil {
				if pageErr := ui.RunTextTUI("搜索失败", err.Error()); pageErr != nil {
					ui.Error("搜索失败: %v", err)
					os.Exit(1)
				}
			}
			keyword = ""
		case ui.MainMenuActionSourceList:
			if err := ui.RunTextTUI("媒体源", renderSourceOverview(application)); err != nil {
				ui.Error("打开媒体源页面失败: %v", err)
				os.Exit(1)
			}
		case ui.MainMenuActionConfigHelp:
			if err := ui.RunTextTUI("配置说明", (&ConfigCommand{}).Usage()); err != nil {
				ui.Error("打开配置说明失败: %v", err)
				os.Exit(1)
			}
		case ui.MainMenuActionVersion:
			if err := ui.RunTextTUI("版本信息", version.Info()); err != nil {
				ui.Error("打开版本信息失败: %v", err)
				os.Exit(1)
			}
		case ui.MainMenuActionQuit:
			return
		default:
			return
		}
	}
}

// Usage 返回使用说明
func (c *TUICommand) Usage() string {
	return "用法: goani tui [keyword]\n\n进入交互式 TUI 模式。\n首页可查看搜索、媒体源、配置说明和版本信息。\n如果提供关键词，会在进入搜索页时自动带入。\n\n示例:\n  goani tui\n  goani tui 葬送的芙莉莲"
}

func renderSourceOverview(application *app.App) string {
	var b strings.Builder

	sources := application.SourceManager.GetAll()
	subs := application.SourceManager.GetSubscriptions()

	b.WriteString(fmt.Sprintf("已加载媒体源: %d\n", len(sources)))
	b.WriteString("\n")
	for i, s := range sources {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, s.Arguments.Name))
		if strings.TrimSpace(s.Arguments.Description) != "" {
			b.WriteString(fmt.Sprintf("   %s\n", s.Arguments.Description))
		}
	}

	b.WriteString("\n订阅列表:\n")
	if len(subs) == 0 {
		b.WriteString("暂无订阅\n")
		return b.String()
	}

	for i, sub := range subs {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, sub.Name))
		b.WriteString(fmt.Sprintf("   URL: %s\n", sub.URL))
		if strings.TrimSpace(sub.UpdatedAt) != "" {
			b.WriteString(fmt.Sprintf("   更新时间: %s\n", sub.UpdatedAt))
		}
	}

	return b.String()
}
