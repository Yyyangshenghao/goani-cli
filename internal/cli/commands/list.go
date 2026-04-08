package commands

import (
	"fmt"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/ui"
)

func init() {
	Register(&ListCommand{app: app.New()})
}

// ListCommand 列表命令
type ListCommand struct {
	app *app.App
}

// Name 返回命令名称
func (c *ListCommand) Name() string {
	return "list"
}

// ShortDesc 返回简短描述
func (c *ListCommand) ShortDesc() string {
	return "列出所有媒体源"
}

// Run 执行命令
func (c *ListCommand) Run(args []string) {
	sources := c.app.SourceManager.GetAll()
	ui.Info("共 %d 个媒体源", len(sources))
	fmt.Println()
	for i, s := range sources {
		fmt.Printf("  %d. %s\n", i+1, s.Arguments.Name)
	}
}

// Usage 返回使用说明
func (c *ListCommand) Usage() string {
	return "用法: goani list"
}
