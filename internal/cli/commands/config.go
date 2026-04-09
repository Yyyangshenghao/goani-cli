package commands

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/ui"
)

func init() {
	Register(&ConfigCommand{})
}

// ConfigCommand 配置命令
type ConfigCommand struct {
	app *app.App
}

func (c *ConfigCommand) ensureApp() *app.App {
	if c.app == nil {
		c.app = app.New()
	}
	return c.app
}

// Name 返回命令名称
func (c *ConfigCommand) Name() string {
	return "config"
}

// ShortDesc 返回简短描述
func (c *ConfigCommand) ShortDesc() string {
	return "配置播放器与其他设置"
}

// Run 执行命令
func (c *ConfigCommand) Run(args []string) {
	if len(args) < 2 || args[0] != "player" {
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	if args[1] == "default" {
		c.setDefaultPlayer(args[2:])
		return
	}

	c.setPlayerPath(args[1:])
}

func (c *ConfigCommand) setPlayerPath(args []string) {
	if len(args) < 2 {
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	name := args[0]
	path := args[1]
	if !player.IsSupportedPlayer(name) {
		ui.Error("不支持的播放器: %s", name)
		os.Exit(1)
	}

	application := c.ensureApp()
	application.PlayerConfig.SetPlayer(name, path)
	if err := application.SaveConfig(); err != nil {
		ui.Error("保存配置失败: %v", err)
		os.Exit(1)
	}

	ui.Success("已配置播放器: %s", name)
}

func (c *ConfigCommand) setDefaultPlayer(args []string) {
	if len(args) < 1 {
		fmt.Println(c.Usage())
		os.Exit(1)
	}

	name := args[0]
	if !player.IsSupportedPlayer(name) {
		ui.Error("不支持的播放器: %s", name)
		os.Exit(1)
	}

	application := c.ensureApp()
	application.PlayerConfig.SetDefaultPlayer(name)
	if err := application.SaveConfig(); err != nil {
		ui.Error("保存配置失败: %v", err)
		os.Exit(1)
	}

	ui.Success("已设置默认播放器: %s", name)
}

// Usage 返回使用说明
func (c *ConfigCommand) Usage() string {
	return `用法:
  goani config player <name> <path>
  goani config player default <name>

示例:
  Windows:   goani config player mpv "D:\MPV播放器\mpv.exe"
  设置默认:   goani config player default mpv
  macOS:     goani config player iina "/Applications/IINA.app/Contents/MacOS/iina-cli"
  Linux:     goani config player mpv "/usr/bin/mpv"`
}
