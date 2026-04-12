package commands

import (
	"fmt"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	consoleui "github.com/Yyyangshenghao/goani-cli/internal/ui/console"
)

func init() {
	Register(&SourceCommand{})
}

// SourceCommand 媒体源命令
type SourceCommand struct {
	app *app.App
}

func (c *SourceCommand) ensureApp() *app.App {
	if c.app == nil {
		c.app = app.New()
	}
	return c.app
}

// Name 返回命令名称
func (c *SourceCommand) Name() string {
	return "source"
}

// ShortDesc 返回简短描述
func (c *SourceCommand) ShortDesc() string {
	return "管理媒体源订阅、渠道开关和探测"
}

// Run 执行命令
func (c *SourceCommand) Run(args []string) {
	if len(args) < 1 {
		fmt.Println(c.Usage())
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "list":
		c.listSources()
	case "doctor":
		c.doctor()
	case "sub":
		c.subscribe(args[1:])
	case "unsub":
		c.unsubscribe(args[1:])
	case "refresh":
		c.refresh()
	case "reset":
		c.reset()
	default:
		fmt.Printf("未知子命令: %s\n", subCmd)
		fmt.Println(c.Usage())
	}
}

func (c *SourceCommand) listSources() {
	application := c.ensureApp()
	channels := application.SourceManager.ListChannels()
	subs := application.SourceManager.GetSubscriptions()
	enabledCount := application.SourceManager.EnabledCount()

	consoleui.Info("共 %d 个媒体源，当前启用 %d 个", len(channels), enabledCount)
	fmt.Printf("  渠道偏好文件: %s\n", source.GetPreferencesPath())
	fmt.Println()

	for i, item := range channels {
		fmt.Printf("  %d. [%s] %s\n", i+1, enabledFlag(item.Enabled), item.Name)
		if strings.TrimSpace(item.Description) != "" {
			fmt.Printf("     %s\n", item.Description)
		}
		if item.Priority > 0 {
			fmt.Printf("     优先级: %d\n", item.Priority)
		}
		if item.LastDoctor != nil {
			fmt.Printf("     最近诊断: %s | %s\n", doctorSummaryLabel(item.LastDoctor), item.LastDoctor.CheckedAt)
			if strings.TrimSpace(item.LastDoctor.Summary) != "" {
				fmt.Printf("     结果说明: %s\n", item.LastDoctor.Summary)
			}
		}
	}

	if len(subs) > 0 {
		fmt.Println()
		consoleui.Info("订阅列表:")
		for i, sub := range subs {
			fmt.Printf("  %d. %s (%s)\n", i+1, sub.Name, sub.URL)
			fmt.Printf("     更新时间: %s\n", sub.UpdatedAt)
		}
	}
}

func (c *SourceCommand) doctor() {
	application := c.ensureApp()
	consoleui.Info("正在执行媒体源 doctor，固定样本: 葬送的芙莉莲 / 进击的巨人 / 鬼灭之刃 / 海贼王 / 名侦探柯南")

	results, err := application.RunSourceDoctor()
	if err != nil {
		consoleui.Error("doctor 执行失败: %v", err)
		return
	}

	for _, result := range results {
		state := "保留"
		if !result.Enabled {
			state = "关闭"
		}
		fmt.Printf("  [%s] %s  成功 %d/%d  优先级 %d  平均 %dms\n", state, result.Name, result.Snapshot.SuccessfulRuns, result.Snapshot.TotalRuns, result.Priority, result.Snapshot.AverageMS)
		fmt.Printf("     %s\n", result.Snapshot.Summary)
	}

	enabledCount := application.SourceManager.EnabledCount()
	fmt.Println()
	consoleui.Success("doctor 完成：当前保留 %d/%d 个源，排序已按诊断结果更新", enabledCount, len(results))
}

func (c *SourceCommand) subscribe(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: goani source sub <url> [name]")
		return
	}

	url := args[0]
	name := "自定义源"
	if len(args) > 1 {
		name = args[1]
	}

	consoleui.Info("正在订阅: %s", url)

	application := c.ensureApp()
	if err := application.SourceManager.Subscribe(url, name); err != nil {
		consoleui.Error("订阅失败: %v", err)
		return
	}

	consoleui.Success("订阅成功！当前共 %d 个媒体源", application.SourceManager.Count())
}

func (c *SourceCommand) unsubscribe(args []string) {
	if len(args) < 1 {
		fmt.Println("用法: goani source unsub <url>")
		return
	}

	url := args[0]

	if err := c.ensureApp().SourceManager.Unsubscribe(url); err != nil {
		consoleui.Error("取消订阅失败: %v", err)
		return
	}

	consoleui.Success("已取消订阅")
}

func (c *SourceCommand) refresh() {
	consoleui.Info("正在刷新订阅...")

	application := c.ensureApp()
	if err := application.SourceManager.Refresh(); err != nil {
		consoleui.Error("刷新失败: %v", err)
		return
	}

	consoleui.Success("刷新完成！当前共 %d 个媒体源", application.SourceManager.Count())
}

func (c *SourceCommand) reset() {
	if err := c.ensureApp().SourceManager.Reset(); err != nil {
		consoleui.Error("重置失败: %v", err)
		return
	}

	consoleui.Success("已重置为默认媒体源")
}

// Usage 返回使用说明
func (c *SourceCommand) Usage() string {
	return `用法:
  goani source list                                列出所有媒体源与渠道状态
  goani source doctor                              用固定 5 个动漫样本诊断媒体源
  goani source sub <url> [name]                    订阅媒体源
  goani source unsub <url>                         取消订阅
  goani source refresh                             刷新订阅
  goani source reset                               重置为默认源

示例:
  goani source sub https://example.com/sources.json 我的订阅
  goani source list
  goani source doctor
  goani source refresh`
}

func enabledFlag(enabled bool) string {
	if enabled {
		return "开"
	}
	return "关"
}

func doctorSummaryLabel(snapshot *source.SourceDoctorSnapshot) string {
	if snapshot == nil {
		return "未诊断"
	}
	return fmt.Sprintf("%d/%d 成功", snapshot.SuccessfulRuns, snapshot.TotalRuns)
}
