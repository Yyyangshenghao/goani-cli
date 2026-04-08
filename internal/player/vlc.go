package player

import (
	"os"
	"os/exec"
	"runtime"
)

// VLCPlayer vlc 播放器
type VLCPlayer struct {
	path string
}

// NewVLCPlayer 创建 vlc 播放器
func NewVLCPlayer() *VLCPlayer {
	path, _ := exec.LookPath("vlc")
	if path == "" && runtime.GOOS == "windows" {
		paths := []string{
			`C:\Program Files\VideoLAN\VLC\vlc.exe`,
			`C:\Program Files (x86)\VideoLAN\VLC\vlc.exe`,
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}
	return &VLCPlayer{path: path}
}

// Name 返回播放器名称
func (p *VLCPlayer) Name() string {
	return "vlc"
}

// Play 播放视频
func (p *VLCPlayer) Play(url string) error {
	return p.PlayWithArgs(url, nil)
}

// PlayWithArgs 带参数播放
func (p *VLCPlayer) PlayWithArgs(url string, args []string) error {
	cmdArgs := append([]string{url}, args...)
	cmd := exec.Command(p.path, cmdArgs...)
	return cmd.Start()
}

// IsAvailable 检查播放器是否可用
func (p *VLCPlayer) IsAvailable() bool {
	return p.path != ""
}
