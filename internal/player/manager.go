package player

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
	return NewManager()
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
