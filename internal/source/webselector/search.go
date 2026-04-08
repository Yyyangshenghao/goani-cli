package webselector

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
)

// Search 搜索动漫
func (s *WebSelectorSource) Search(keyword string) ([]source.Anime, error) {
	searchURL := s.config.Arguments.SearchConfig.SearchURL
	searchURL = strings.ReplaceAll(searchURL, "{keyword}", keyword)

	doc, err := s.fetchHTML(searchURL)
	if err != nil {
		return nil, err
	}

	var results []source.Anime
	searchConfig := s.config.Arguments.SearchConfig

	switch searchConfig.SubjectFormatID {
	case "a":
		results = s.parseSubjectFormatA(doc, searchConfig)
	case "indexed":
		results = s.parseSubjectFormatIndexed(doc, searchConfig)
	}

	return results, nil
}

// parseSubjectFormatA 解析格式 A 的搜索结果
func (s *WebSelectorSource) parseSubjectFormatA(doc *goquery.Document, config source.SearchConfig) []source.Anime {
	var results []source.Anime
	selector := config.SelectorSubjectFormatA.SelectLists

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Text())
		href, exists := sel.Attr("href")
		if exists && name != "" {
			results = append(results, source.Anime{
				Name: name,
				URL:  s.resolveURL(href),
			})
		}
	})

	return results
}

// parseSubjectFormatIndexed 解析索引格式的搜索结果
func (s *WebSelectorSource) parseSubjectFormatIndexed(doc *goquery.Document, config source.SearchConfig) []source.Anime {
	var results []source.Anime

	names := doc.Find(config.SelectorSubjectFormatIndexed.SelectNames)
	links := doc.Find(config.SelectorSubjectFormatIndexed.SelectLinks)

	for i := range names.Nodes {
		if i >= links.Length() {
			break
		}
		name := strings.TrimSpace(names.Eq(i).Text())
		href, exists := links.Eq(i).Attr("href")
		if exists && name != "" {
			results = append(results, source.Anime{
				Name: name,
				URL:  s.resolveURL(href),
			})
		}
	}

	return results
}
