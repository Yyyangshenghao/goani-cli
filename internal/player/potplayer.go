package player

import (
	"os"
	"os/exec"
	"runtime"
)

// PotPlayer PotPlayer 播放器 (Windows)
type PotPlayer struct {
	path string
}

// NewPotPlayer 创建 PotPlayer 播放器
func NewPotPlayer() *PotPlayer {
	path, _ := exec.LookPath("PotPlayerMini64")
	if path == "" {
		path, _ = exec.LookPath("PotPlayerMini")
	}
	if path == "" && runtime.GOOS == "windows" {
		paths := []string{
			`C:\Program Files\DAUM\PotPlayer\PotPlayerMini64.exe`,
			`C:\Program Files (x86)\DAUM\PotPlayer\PotPlayerMini.exe`,
			`C:\Program Files\DAUM\PotPlayer\PotPlayerMini.exe`,
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}
	return &PotPlayer{path: path}
}

// Name 返回播放器名称
func (p *PotPlayer) Name() string {
	return "potplayer"
}

// Play 播放视频
func (p *PotPlayer) Play(url string) error {
	return p.PlayWithArgs(url, nil)
}

// PlayWithArgs 带参数播放
func (p *PotPlayer) PlayWithArgs(url string, args []string) error {
	cmdArgs := append(append([]string{}, args...), url)
	cmd := exec.Command(p.path, cmdArgs...)
	return cmd.Start()
}

// IsAvailable 检查播放器是否可用
func (p *PotPlayer) IsAvailable() bool {
	return p.path != ""
}
