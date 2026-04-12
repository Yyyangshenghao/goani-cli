package webselector

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/dlclark/regexp2"
)

// GetEpisodes 获取剧集列表
func (s *WebSelectorSource) GetEpisodes(animeURL string) ([]source.Episode, error) {
	doc, err := s.fetchHTML(animeURL)
	if err != nil {
		return nil, err
	}

	var episodes []source.Episode
	config := s.config.Arguments.SearchConfig

	switch config.ChannelFormatID {
	case "index-grouped":
		episodes = s.parseChannelFormatFlattened(doc, config)
	case "no-channel":
		episodes = s.parseChannelFormatNoChannel(doc, config)
	}

	return episodes, nil
}

// parseChannelFormatFlattened 解析扁平化频道的剧集
func (s *WebSelectorSource) parseChannelFormatFlattened(doc *goquery.Document, config source.SearchConfig) []source.Episode {
	var episodes []source.Episode
	selector := config.SelectorChannelFormatFlattened

	lists := doc.Find(selector.SelectEpisodeLists)
	lists.Each(func(i int, list *goquery.Selection) {
		list.Find(selector.SelectEpisodesFromList).Each(func(j int, ep *goquery.Selection) {
			name := strings.TrimSpace(ep.Text())
			href, exists := ep.Attr("href")
			if exists && name != "" {
				number, numberValue, hasNumber := parseEpisodeNumber(name, selector.MatchEpisodeSortFromName)
				episodes = append(episodes, source.Episode{
					Name:        name,
					URL:         s.resolveURL(href),
					Number:      number,
					NumberValue: numberValue,
					HasNumber:   hasNumber,
				})
			}
		})
	})

	return episodes
}

// parseChannelFormatNoChannel 解析无频道的剧集
func (s *WebSelectorSource) parseChannelFormatNoChannel(doc *goquery.Document, config source.SearchConfig) []source.Episode {
	var episodes []source.Episode
	selector := config.SelectorChannelFormatNoChannel

	doc.Find(selector.SelectEpisodes).Each(func(i int, ep *goquery.Selection) {
		name := strings.TrimSpace(ep.Text())
		href, exists := ep.Attr("href")
		if exists && name != "" {
			number, numberValue, hasNumber := parseEpisodeNumber(name, selector.MatchEpisodeSortFromName)
			episodes = append(episodes, source.Episode{
				Name:        name,
				URL:         s.resolveURL(href),
				Number:      number,
				NumberValue: numberValue,
				HasNumber:   hasNumber,
			})
		}
	})

	return episodes
}

// parseEpisodeNumber 先走源配置里的集数规则，失败后再用通用数字提取兜底。
func parseEpisodeNumber(name, pattern string) (string, float64, bool) {
	if looksLikeEpisodeRange(name) {
		return "", 0, false
	}

	if value := extractEpisodeNumberWithPattern(name, pattern); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return strconv.FormatFloat(parsed, 'f', -1, 64), parsed, true
		}
	}

	re := regexp2.MustCompile(`(?<ep>\d+(?:\.\d+)?)`, regexp2.None)
	match, err := re.FindStringMatch(name)
	if err != nil || match == nil {
		return "", 0, false
	}

	group := match.GroupByName("ep")
	if strings.TrimSpace(group.String()) == "" {
		return "", 0, false
	}

	parsed, err := strconv.ParseFloat(group.String(), 64)
	if err != nil {
		return "", 0, false
	}

	return strconv.FormatFloat(parsed, 'f', -1, 64), parsed, true
}

func looksLikeEpisodeRange(name string) bool {
	re := regexp2.MustCompile(`\d+\s*[-~～—]+\s*\d+`, regexp2.None)
	match, err := re.FindStringMatch(name)
	return err == nil && match != nil
}

// extractEpisodeNumberWithPattern 兼容 animeko-source 这类配置里的正则规则。
func extractEpisodeNumberWithPattern(name, pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return ""
	}

	re, err := regexp2.Compile(pattern, regexp2.None)
	if err != nil {
		return ""
	}

	match, err := re.FindStringMatch(name)
	if err != nil || match == nil {
		return ""
	}

	if group := match.GroupByName("ep"); strings.TrimSpace(group.String()) != "" {
		return strings.TrimSpace(group.String())
	}

	if match.GroupCount() >= 2 {
		group := match.Groups()[1]
		if strings.TrimSpace(group.String()) != "" {
			return strings.TrimSpace(group.String())
		}
	}

	return strings.TrimSpace(match.String())
}
