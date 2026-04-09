package player

var supportedPlayers = []string{"mpv", "vlc", "potplayer", "iina"}

// Manager 播放器管理器
type Manager struct {
	players []Player
}

// NewManager 创建播放器管理器
func NewManager() *Manager {
	return &Manager{players: buildPlayers("", nil)}
}

// NewManagerWithConfig 使用配置创建播放器管理器
func NewManagerWithConfig(defaultPlayer string, paths map[string]string) *Manager {
	return &Manager{players: buildPlayers(defaultPlayer, paths)}
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

// IsSupportedPlayer 是否为支持的播放器
func IsSupportedPlayer(name string) bool {
	for _, supported := range supportedPlayers {
		if name == supported {
			return true
		}
	}
	return false
}

func buildPlayers(defaultPlayer string, paths map[string]string) []Player {
	order := orderedPlayers(defaultPlayer)
	players := make([]Player, 0, len(order))
	for _, name := range order {
		if p := newPlayer(name, paths[name]); p != nil {
			players = append(players, p)
		}
	}
	return players
}

func orderedPlayers(defaultPlayer string) []string {
	order := make([]string, 0, len(supportedPlayers))
	if defaultPlayer != "" && IsSupportedPlayer(defaultPlayer) {
		order = append(order, defaultPlayer)
	}
	for _, name := range supportedPlayers {
		if name != defaultPlayer {
			order = append(order, name)
		}
	}
	return order
}

func newPlayer(name, path string) Player {
	switch name {
	case "mpv":
		if path != "" {
			return NewMPVPlayerWithPath(path)
		}
		return NewMPVPlayer()
	case "vlc":
		if path != "" {
			return &VLCPlayer{path: path}
		}
		return NewVLCPlayer()
	case "potplayer":
		if path != "" {
			return &PotPlayer{path: path}
		}
		return NewPotPlayer()
	case "iina":
		if path != "" {
			return &IINAPlayer{path: path}
		}
		return NewIINAPlayer()
	default:
		return nil
	}
}
