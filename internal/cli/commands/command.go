package commands

// Command 命令接口
type Command interface {
	// Name 返回命令名称
	Name() string
	// Run 执行命令
	Run(args []string)
	// Usage 返回使用说明
	Usage() string
}
