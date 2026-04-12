package commands

import (
	"fmt"
	"os"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
)

func init() {
	Register(&BridgeHLSCommand{})
}

// BridgeHLSCommand 是给播放器兼容层使用的内部命令，不对普通用户暴露。
type BridgeHLSCommand struct{}

func (c *BridgeHLSCommand) Name() string {
	return "bridge-hls"
}

func (c *BridgeHLSCommand) ShortDesc() string {
	return "内部 ffmpeg HLS 桥接"
}

func (c *BridgeHLSCommand) Usage() string {
	return "用法: goani bridge-hls <encoded-stream-context>"
}

func (c *BridgeHLSCommand) Hidden() bool {
	return true
}

func (c *BridgeHLSCommand) Run(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, c.Usage())
		os.Exit(1)
	}

	ctx, err := player.DecodeStreamRequestContext(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := player.ServeFFmpegHLSBridge(ctx, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
