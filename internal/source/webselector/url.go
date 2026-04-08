package webselector

import "strings"

// resolveURL 解析相对 URL
func (s *WebSelectorSource) resolveURL(href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}

	// 从搜索 URL 提取 base URL
	searchURL := s.config.Arguments.SearchConfig.SearchURL
	idx := strings.Index(searchURL, "://")
	if idx != -1 {
		rest := searchURL[idx+3:]
		slashIdx := strings.Index(rest, "/")
		if slashIdx != -1 {
			base := searchURL[:idx+3+slashIdx]
			if strings.HasPrefix(href, "/") {
				return base + href
			}
			return base + "/" + href
		}
	}
	return href
}
