package player

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
)

// Player 播放器接口
type Player interface {
    // Name 返回播放器名称
    Name() string
    // Play 播放视频
    Play(url string) error
    // PlayWithArgs 带参数播放
    PlayWithArgs(url string, args []string) error
    // IsAvailable 检查播放器是否可用
    IsAvailable() bool
}

// MPVPlayer mpv 播放器
type MPVPlayer struct {
    path string
}

// NewMPVPlayer 创建 mpv 播放器
func NewMPVPlayer() *MPVPlayer {
    path, _ := exec.LookPath("mpv")
    if path == "" && runtime.GOOS == "windows" {
        // Windows 常见安装路径
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

// VLCPlayer vlc 播放器
type VLCPlayer struct {
    path string
}

// NewVLCPlayer 创建 vlc 播放器
func NewVLCPlayer() *VLCPlayer {
    path, _ := exec.LookPath("vlc")
    if path == "" && runtime.GOOS == "windows" {
        // Windows 常见安装路径
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
    cmdArgs := append([]string{url}, args...)
    cmd := exec.Command(p.path, cmdArgs...)
    return cmd.Start()
}

// IsAvailable 检查播放器是否可用
func (p *PotPlayer) IsAvailable() bool {
    return p.path != ""
}

// IINAPlayer IINA 播放器 (macOS)
type IINAPlayer struct {
    path string
}

// NewIINAPlayer 创建 IINA 播放器
func NewIINAPlayer() *IINAPlayer {
    path, _ := exec.LookPath("iina")
    if path == "" && runtime.GOOS == "darwin" {
        // macOS 应用路径
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
    cmdArgs := append([]string{url}, args...)
    cmd := exec.Command(p.path, cmdArgs...)
    return cmd.Start()
}

// IsAvailable 检查播放器是否可用
func (p *IINAPlayer) IsAvailable() bool {
    return p.path != ""
}

// Manager 播放器管理器
type Manager struct {
    players []Player
}

// NewManager 创建播放器管理器
func NewManager() *Manager {
    return &Manager{
        players: []Player{
            NewMPVPlayer(),
            NewVLCPlayer(),
            NewPotPlayer(),
            NewIINAPlayer(),
        },
    }
}

// NewManagerWithConfig 使用配置创建播放器管理器
func NewManagerWithConfig(playerName, playerPath string) *Manager {
    m := NewManager()

    // 如果配置了自定义路径，优先使用
    if playerPath != "" {
        switch playerName {
        case "mpv":
            return &Manager{players: []Player{NewMPVPlayerWithPath(playerPath)}}
        case "vlc":
            return &Manager{players: []Player{&VLCPlayer{path: playerPath}}}
        case "potplayer":
            return &Manager{players: []Player{&PotPlayer{path: playerPath}}}
        case "iina":
            return &Manager{players: []Player{&IINAPlayer{path: playerPath}}}
        }
    }
    }

    return m
}

// GetAvailable 获取所有可用的播放器
func (m *Manager) GetAvailable() []Player {
    var available []Player
    for _, p := range m.players {
        if p.IsAvailable() {
            available = append(available, p)
        }
    }
    return available
}

// GetFirst 获取第一个可用的播放器
func (m *Manager) GetFirst() Player {
    for _, p := range m.players {
        if p.IsAvailable() {
            return p
        }
    }
    return nil
}

// GetByName 根据名称获取播放器
func (m *Manager) GetByName(name string) Player {
    for _, p := range m.players {
        if p.Name() == name && p.IsAvailable() {
            return p
        }
    }
    return nil
}
