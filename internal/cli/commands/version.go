package commands

import "fmt"

// VersionCommand 版本命令
type VersionCommand struct {
	version string
}

// Name 返回命令名称
func (c *VersionCommand) Name() string {
	return "version"
}

// Run 执行命令
func (c *VersionCommand) Run(args []string) {
	fmt.Printf("goani v%s\n", c.version)
}

// Usage 返回使用说明
func (c *VersionCommand) Usage() string {
	return "用法: goani version"
}

// SetVersion 设置版本号（需要在 main 中调用）
func SetVersion(v string) {
	Register(&VersionCommand{version: v})
}
