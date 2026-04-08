package player

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// MPVPlayer mpv 播放器
type MPVPlayer struct {
	path string
}

// NewMPVPlayer 创建 mpv 播放器
func NewMPVPlayer() *MPVPlayer {
	path, _ := exec.LookPath("mpv")
	if path == "" && runtime.GOOS == "windows" {
		paths := []string{
			`C:\Program Files\mpv\mpv.exe`,
			`C:\Program Files (x86)\mpv\mpv.exe`,
			filepath.Join(os.Getenv("APPDATA"), "mpv", "mpv.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "mpv", "mpv.exe"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}
	return &MPVPlayer{path: path}
}

// NewMPVPlayerWithPath 使用自定义路径创建 mpv 播放器
func NewMPVPlayerWithPath(path string) *MPVPlayer {
	return &MPVPlayer{path: path}
}

// Name 返回播放器名称
func (p *MPVPlayer) Name() string {
	return "mpv"
}

// Play 播放视频
func (p *MPVPlayer) Play(url string) error {
	return p.PlayWithArgs(url, nil)
}

// PlayWithArgs 带参数播放
func (p *MPVPlayer) PlayWithArgs(url string, args []string) error {
	cmdArgs := append([]string{url}, args...)
	cmd := exec.Command(p.path, cmdArgs...)
	return cmd.Start()
}

// IsAvailable 检查播放器是否可用
func (p *MPVPlayer) IsAvailable() bool {
	return p.path != ""
}
