package source

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dlclark/regexp2"
)

// WebSelectorSource 基于 CSS 选择器的媒体源实现
type WebSelectorSource struct {
	config MediaSource
	client *http.Client
}

// NewWebSelectorSource 创建新的媒体源
func NewWebSelectorSource(config MediaSource) *WebSelectorSource {
	return &WebSelectorSource{
		config: config,
		client: &http.Client{},
	}
}

// Name 返回源名称
func (s *WebSelectorSource) Name() string {
	return s.config.Arguments.Name
}

// Search 搜索动漫
func (s *WebSelectorSource) Search(keyword string) ([]Anime, error) {
	searchURL := s.config.Arguments.SearchConfig.SearchURL
	searchURL = strings.ReplaceAll(searchURL, "{keyword}", keyword)

	doc, err := s.fetchHTML(searchURL)
	if err != nil {
		return nil, err
	}

	var results []Anime
	searchConfig := s.config.Arguments.SearchConfig

	// 根据 subjectFormatId 选择不同的解析方式
	switch searchConfig.SubjectFormatID {
	case "a":
		results = s.parseSubjectFormatA(doc, searchConfig)
	case "indexed":
		results = s.parseSubjectFormatIndexed(doc, searchConfig)
	}

	return results, nil
}

// parseSubjectFormatA 解析格式 A 的搜索结果
func (s *WebSelectorSource) parseSubjectFormatA(doc *goquery.Document, config SearchConfig) []Anime {
	var results []Anime
	selector := config.SelectorSubjectFormatA.SelectLists

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Text())
		href, exists := sel.Attr("href")
		if exists && name != "" {
			results = append(results, Anime{
				Name: name,
				URL:  s.resolveURL(href),
			})
		}
	})

	return results
}

// parseSubjectFormatIndexed 解析索引格式的搜索结果
func (s *WebSelectorSource) parseSubjectFormatIndexed(doc *goquery.Document, config SearchConfig) []Anime {
	var results []Anime

	names := doc.Find(config.SelectorSubjectFormatIndexed.SelectNames)
	links := doc.Find(config.SelectorSubjectFormatIndexed.SelectLinks)

	for i := range names.Nodes {
		if i >= links.Length() {
			break
		}
		name := strings.TrimSpace(names.Eq(i).Text())
		href, exists := links.Eq(i).Attr("href")
		if exists && name != "" {
			results = append(results, Anime{
				Name: name,
				URL:  s.resolveURL(href),
			})
		}
	}

	return results
}

// GetEpisodes 获取剧集列表
func (s *WebSelectorSource) GetEpisodes(animeURL string) ([]Episode, error) {
	doc, err := s.fetchHTML(animeURL)
	if err != nil {
		return nil, err
	}

	var episodes []Episode
	config := s.config.Arguments.SearchConfig

	// 根据 channelFormatId 选择不同的解析方式
	switch config.ChannelFormatID {
	case "index-grouped":
		episodes = s.parseChannelFormatFlattened(doc, config)
	case "no-channel":
		episodes = s.parseChannelFormatNoChannel(doc, config)
	}

	return episodes, nil
}

// parseChannelFormatFlattened 解析扁平化频道的剧集
func (s *WebSelectorSource) parseChannelFormatFlattened(doc *goquery.Document, config SearchConfig) []Episode {
	var episodes []Episode
	selector := config.SelectorChannelFormatFlattened

	// 找到所有剧集列表
	lists := doc.Find(selector.SelectEpisodeLists)
	lists.Each(func(i int, list *goquery.Selection) {
		list.Find(selector.SelectEpisodesFromList).Each(func(j int, ep *goquery.Selection) {
			name := strings.TrimSpace(ep.Text())
			href, exists := ep.Attr("href")
			if exists && name != "" {
				episodes = append(episodes, Episode{
					Name: name,
					URL:  s.resolveURL(href),
				})
			}
		})
	})

	return episodes
}

// parseChannelFormatNoChannel 解析无频道的剧集
func (s *WebSelectorSource) parseChannelFormatNoChannel(doc *goquery.Document, config SearchConfig) []Episode {
	var episodes []Episode
	selector := config.SelectorChannelFormatNoChannel

	doc.Find(selector.SelectEpisodes).Each(func(i int, ep *goquery.Selection) {
		name := strings.TrimSpace(ep.Text())
		href, exists := ep.Attr("href")
		if exists && name != "" {
			episodes = append(episodes, Episode{
				Name: name,
				URL:  s.resolveURL(href),
			})
		}
	})

	return episodes
}

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
	// 查找 player_aaaa 变量
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

// fetchHTML 获取 HTML 文档
func (s *WebSelectorSource) fetchHTML(url string) (*goquery.Document, error) {
	body, err := s.fetchBytes(url)
	if err != nil {
		return nil, err
	}

	// 尝试解码 JSON 字符串
	var htmlStr string
	if err := json.Unmarshal(body, &htmlStr); err == nil {
		body = []byte(htmlStr)
	}

	return goquery.NewDocumentFromReader(strings.NewReader(string(body)))
}

// fetchBytes 获取原始字节
func (s *WebSelectorSource) fetchBytes(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// fetchRawHTML 获取原始 HTML 字符串
func (s *WebSelectorSource) fetchRawHTML(url string) (string, error) {
	body, err := s.fetchBytes(url)
	if err != nil {
		return "", err
	}

	// 尝试解码 JSON 字符串
	var htmlStr string
	if err := json.Unmarshal(body, &htmlStr); err == nil {
		return htmlStr, nil
	}
	return string(body), nil
}

// resolveURL 解析相对 URL
func (s *WebSelectorSource) resolveURL(href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	// 从搜索 URL 提取 base URL
	searchURL := s.config.Arguments.SearchConfig.SearchURL
	// 简单提取：找到 "://" 后第一个 "/" 之前的部分
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

// Client 返回 HTTP 客户端（用于调试）
func (s *WebSelectorSource) Client() *http.Client {
	return s.client
}

// FetchRawHTML 公开的方法，用于调试
func (s *WebSelectorSource) FetchRawHTML(url string) (string, error) {
	return s.fetchRawHTML(url)
}
