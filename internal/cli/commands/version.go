package commands

import (
	"github.com/Yyyangshenghao/goani-cli/internal/version"
)

func init() {
	Register(&VersionCommand{})
}

// VersionCommand 版本命令
type VersionCommand struct{}

// Name 返回命令名称
func (c *VersionCommand) Name() string {
	return "version"
}

// ShortDesc 返回简短描述
func (c *VersionCommand) ShortDesc() string {
	return "显示版本信息"
}

// Run 执行命令
func (c *VersionCommand) Run(args []string) {
	print(version.Info())
}

// Usage 返回使用说明
func (c *VersionCommand) Usage() string {
	return "用法: goani version"
}
