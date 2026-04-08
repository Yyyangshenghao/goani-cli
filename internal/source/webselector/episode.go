package webselector

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
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
				episodes = append(episodes, source.Episode{
					Name: name,
					URL:  s.resolveURL(href),
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
			episodes = append(episodes, source.Episode{
				Name: name,
				URL:  s.resolveURL(href),
			})
		}
	})

	return episodes
}
