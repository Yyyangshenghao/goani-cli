package commands

import (
	"fmt"
	"os"

	"github.com/yshscpu/goani-cli/internal/app"
	"github.com/yshscpu/goani-cli/internal/ui"
)

func init() {
	Register(&ConfigCommand{app: app.New()})
}

// ConfigCommand 配置命令
type ConfigCommand struct {
	app *app.App
}

// Name 返回命令名称
func (c *ConfigCommand) Name() string {
	return "config"
}

// Run 执行命令
func (c *ConfigCommand) Run(args []string) {
	if len(args) < 3 || args[0] != "player" {
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	name := args[1]
	path := args[2]

	c.app.Config.SetPlayer(name, path)
	if err := c.app.SaveConfig(); err != nil {
		ui.Error("保存配置失败: %v", err)
		os.Exit(1)
	}

	ui.Success("已配置播放器: %s", name)
}

// Usage 返回使用说明
func (c *ConfigCommand) Usage() string {
	return `用法: goani config player <name> <path>

示例:
  Windows:   goani config player mpv "D:\MPV播放器\mpv.exe"
  macOS:     goani config player iina "/Applications/IINA.app/Contents/MacOS/iina-cli"
  Linux:     goani config player mpv "/usr/bin/mpv"`
}
