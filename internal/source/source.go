package source

// Anime 搜索结果
type Anime struct {
	Name string
	URL  string
}

// Episode 剧集
type Episode struct {
	Name        string
	URL         string
	SourceName  string
	Number      string
	NumberValue float64
	HasNumber   bool
}

// EpisodeCandidate 剧集候选线路
type EpisodeCandidate struct {
	Name       string
	URL        string
	SourceName string
}

// EpisodeGroup 归类后的剧集
type EpisodeGroup struct {
	Name        string
	Number      string
	NumberValue float64
	HasNumber   bool
	Candidates  []EpisodeCandidate
	order       int
}

// Source 媒体源接口
type Source interface {
	// Name 返回源名称
	Name() string
	// Search 搜索动漫
	Search(keyword string) ([]Anime, error)
	// GetEpisodes 获取剧集列表
	GetEpisodes(animeURL string) ([]Episode, error)
	// GetVideoURL 获取视频直链
	GetVideoURL(episodeURL string) (string, error)
}

// Aggregator 聚合器 - 管理多个源
type Aggregator interface {
	// AddSource 添加源
	AddSource(source Source)
	// SearchAll 并行搜索所有源
	SearchAll(keyword string) (map[string][]Anime, error)
	// GetSource 根据名称获取源
	GetSource(name string) Source
}
