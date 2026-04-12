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
	return "用法: goani proxy-hls <encoded-stream-context>"
}

func (c *ProxyHLSCommand) Hidden() bool {
	return true
}

func (c *ProxyHLSCommand) Run(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, c.Usage())
		os.Exit(1)
	}

	ctx, err := player.DecodeStreamRequestContext(args[0])
	if err != nil {
		ctx = player.StreamRequestContext{
			SourceURL: args[0],
		}
		if len(args) > 1 {
			ctx.Referer = args[1]
		}
		if len(args) > 2 {
			ctx.UserAgent = args[2]
		}
	}

	if err := player.ServeHLSProxy(ctx, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
