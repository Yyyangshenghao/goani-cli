package commands

import (
	"context"
	"fmt"
	"os"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/Yyyangshenghao/goani-cli/internal/mcp"
)

func init() {
	Register(&McpCommand{})
}

// McpCommand MCP 服务器命令
type McpCommand struct{}

// Name 返回命令名称
func (c *McpCommand) Name() string {
	return "mcp"
}

// ShortDesc 返回简短描述
func (c *McpCommand) ShortDesc() string {
	return "启动 MCP 服务器（stdio 模式）"
}

// Run 执行命令
func (c *McpCommand) Run(args []string) {
	server := mcp.NewServer()
	if err := server.Run(context.Background(), &mcpgo.StdioTransport{}); err != nil {
		fmt.Fprintf(os.Stderr, "MCP 服务器运行失败: %v\n", err)
		os.Exit(1)
	}
}

// Usage 返回使用说明
func (c *McpCommand) Usage() string {
	return "用法: goani mcp\n\n启动 MCP 服务器，通过 stdin/stdout 与客户端通信。\n\n示例:\n  goani mcp"
}
