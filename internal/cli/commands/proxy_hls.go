package commands

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
)

func init() {
	Register(&ProxyHLSCommand{})
}

// ProxyHLSCommand 是给播放器兼容层使用的内部命令，不对普通用户暴露。
type ProxyHLSCommand struct{}

func (c *ProxyHLSCommand) Name() string {
	return "proxy-hls"
}

func (c *ProxyHLSCommand) ShortDesc() string {
	return "内部 HLS 代理"
}

func (c *ProxyHLSCommand) Usage() string {
	return "用法: goani proxy-hls <source-url> [referer] [user-agent]"
}

func (c *ProxyHLSCommand) Hidden() bool {
	return true
}

func (c *ProxyHLSCommand) Run(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, c.Usage())
		os.Exit(1)
	}

	sourceURL := args[0]
	referer := ""
	userAgent := ""
	if len(args) > 1 {
		referer = args[1]
	}
	if len(args) > 2 {
		userAgent = args[2]
	}

	if err := player.ServeHLSProxy(sourceURL, referer, userAgent, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
