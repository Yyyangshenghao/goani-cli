package commands

// Command 命令接口
type Command interface {
	// Name 返回命令名称
	Name() string
	// Run 执行命令
	Run(args []string)
	// Usage 返回详细使用说明
	Usage() string
	// ShortDesc 返回简短描述（用于帮助列表）
	ShortDesc() string
}

// HiddenCommand 用于标记不应该展示给普通用户的内部命令。
type HiddenCommand interface {
	Hidden() bool
}
