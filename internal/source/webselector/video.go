package webselector

import (
	"encoding/json"
	"strings"

	"github.com/dlclark/regexp2"
)

// GetVideoURL 获取视频直链
func (s *WebSelectorSource) GetVideoURL(episodeURL string) (string, error) {
	html, err := s.fetchRawHTML(episodeURL)
	if err != nil {
		return "", err
	}

	// 先尝试从 player_aaaa 变量中提取
	if url := extractPlayerAAAA(html); url != "" {
		return url, nil
	}

	// 再用正则匹配
	config := s.config.Arguments.SearchConfig.MatchVideo
	re, err := regexp2.Compile(config.MatchVideoURL, regexp2.None)
	if err != nil {
		return "", err
	}

	match, err := re.FindStringMatch(html)
	if err != nil || match == nil {
		return "", nil
	}

	if v := match.GroupByName("v"); v != nil && v.String() != "" {
		return v.String(), nil
	}

	return match.String(), nil
}

// extractPlayerAAAA 从 player_aaaa 变量中提取视频 URL
func extractPlayerAAAA(html string) string {
	idx := strings.Index(html, "player_aaaa=")
	if idx == -1 {
		return ""
	}

	// 提取 JSON 部分（从 { 开始）
	start := idx + len("player_aaaa=")
	jsonStart := strings.Index(html[start:], "{")
	if jsonStart == -1 {
		return ""
	}
	start = start + jsonStart

	// 找到匹配的 }
	depth := 0
	end := -1
	for i, c := range html[start:] {
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				end = start + i + 1
				break
			}
		}
	}
	if end == -1 {
		return ""
	}
	jsonStr := html[start:end]

	// 解析 JSON
	var player struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &player); err == nil && player.URL != "" {
		return player.URL
	}
	return ""
}
