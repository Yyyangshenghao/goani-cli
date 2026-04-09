package player

import (
	"os"
	"os/exec"
	"runtime"
)

// IINAPlayer IINA 播放器 (macOS)
type IINAPlayer struct {
	path string
}

// NewIINAPlayer 创建 IINA 播放器
func NewIINAPlayer() *IINAPlayer {
	path, _ := exec.LookPath("iina")
	if path == "" && runtime.GOOS == "darwin" {
		appPath := "/Applications/IINA.app/Contents/MacOS/iina-cli"
		if _, err := os.Stat(appPath); err == nil {
			path = appPath
		}
	}
	return &IINAPlayer{path: path}
}

// Name 返回播放器名称
func (p *IINAPlayer) Name() string {
	return "iina"
}

// Play 播放视频
func (p *IINAPlayer) Play(url string) error {
	return p.PlayWithArgs(url, nil)
}

// PlayWithArgs 带参数播放
func (p *IINAPlayer) PlayWithArgs(url string, args []string) error {
	// 使用 open 命令打开 IINA，更可靠
	cmdArgs := []string{"-a", "IINA", url}
	cmd := exec.Command("open", cmdArgs...)
	return cmd.Start()
}

// IsAvailable 检查播放器是否可用
func (p *IINAPlayer) IsAvailable() bool {
	return p.path != ""
}
