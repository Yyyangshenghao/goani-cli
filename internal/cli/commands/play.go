package commands

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	consoleui "github.com/Yyyangshenghao/goani-cli/internal/ui/console"
	"github.com/Yyyangshenghao/goani-cli/internal/workflow"
)

func init() {
	Register(&PlayCommand{})
}

// PlayCommand 播放命令
type PlayCommand struct {
	app *app.App
}

func (c *PlayCommand) ensureApp() *app.App {
	if c.app == nil {
		c.app = app.New()
	}
	return c.app
}

// Name 返回命令名称
func (c *PlayCommand) Name() string {
	return "play"
}

// ShortDesc 返回简短描述
func (c *PlayCommand) ShortDesc() string {
	return "搜索并播放动漫"
}

// Run 执行命令
func (c *PlayCommand) Run(args []string) {
	if len(args) < 1 {
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	application := c.ensureApp()
	keyword := args[0]
	src := application.GetFirstSource()
	if src == nil {
		consoleui.Error("未找到媒体源")
		os.Exit(1)
	}

	consoleui.Info("搜索: %s", keyword)
	results, err := src.Search(keyword)
	if err != nil {
		consoleui.Error("搜索失败: %v", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		consoleui.Info("未找到结果")
		return
	}

	// 选择动漫
	idx, err := consoleui.Select("选择动漫", len(results), func(i int) string { return results[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	// 获取剧集
	episodes, err := src.GetEpisodes(results[idx].URL)
	if err != nil {
		consoleui.Error("获取剧集失败: %v", err)
		os.Exit(1)
	}
	groups := source.GroupEpisodes(episodes)
	if len(groups) == 0 {
		consoleui.Info("没有可用剧集")
		return
	}

	consoleui.Success("找到 %d 集", len(groups))

	// 选择剧集
	epIdx, err := consoleui.Select("选择剧集", len(groups), func(i int) string { return groups[i].Label() })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	// 播放
	if err := workflow.PlayEpisodeGroupCLI(application, src, groups[epIdx]); err != nil {
		consoleui.Error("%v", err)
		os.Exit(1)
	}
}

// Usage 返回使用说明
func (c *PlayCommand) Usage() string {
	return "用法: goani play <keyword>\n示例: goani play 葬送的芙莉莲"
}
