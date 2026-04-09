package webselector

import (
	"net/http"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
)

// WebSelectorSource 基于 CSS 选择器的媒体源实现
type WebSelectorSource struct {
	config source.MediaSource
	client *http.Client
}

// New 创建新的媒体源
func New(config source.MediaSource) *WebSelectorSource {
	return &WebSelectorSource{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 返回源名称
func (s *WebSelectorSource) Name() string {
	return s.config.Arguments.Name
}

// Client 返回 HTTP 客户端（用于调试）
func (s *WebSelectorSource) Client() *http.Client {
	return s.client
}

// TestSourceManager 测试用的媒体源管理器
type TestSourceManager struct {
	sources []source.MediaSource
}

// NewTestSourceManager 创建测试用的媒体源管理器
func NewTestSourceManager() *TestSourceManager {
	return &TestSourceManager{
		sources: []source.MediaSource{
			{
				FactoryID: "web-selector",
				Version:   2,
				Arguments: source.Arguments{
					Name:        "omofun111",
					Description: "测试源",
					SearchConfig: source.SearchConfig{
						SearchURL:       "https://enlienli.link/vod/search.html?wd={keyword}",
						SubjectFormatID: "a",
						SelectorSubjectFormatA: source.SelectorSubjectFormatA{
							SelectLists: ".module-card-item>.module-card-item-info>.module-card-item-title>a",
						},
						ChannelFormatID: "index-grouped",
						SelectorChannelFormatFlattened: source.SelectorChannelFormatFlattened{
							SelectEpisodeLists:     ".module-play-list-content",
							SelectEpisodesFromList: "a",
						},
						MatchVideo: source.MatchVideo{
							EnableNestedURL: true,
							MatchNestedURL:  "$^",
							MatchVideoURL:   "(^http(s)?:\\/\\/(?!.*http).+\\.(mp4|m3u8|flv|mkv))|(url=(?<v>http(s)?:\\/\\/.+\\.(mp4|m3u8|flv|mkv)))",
						},
					},
				},
			},
		},
	}
}

// GetAll 获取所有媒体源
func (m *TestSourceManager) GetAll() []source.MediaSource {
	return m.sources
}
