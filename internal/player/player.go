package player

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
