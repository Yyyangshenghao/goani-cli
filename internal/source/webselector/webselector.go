package webselector

import (
	"net/http"

	"github.com/yshscpu/goani-cli/internal/source"
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
		client: &http.Client{},
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
