package workflow

import (
	"fmt"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

func loadAggregatedEpisodeGroupsTUI(application *app.App, anime source.AggregatedAnime) ([]source.EpisodeGroup, error) {
	var groups []source.EpisodeGroup
	err := tui.RunLoadingTUI(
		"正在聚合剧集",
		"正在从多个片源汇总剧集和线路，稍等一下",
		func() error {
			var loadErr error
			groups, loadErr = loadAggregatedEpisodeGroups(application, anime)
			return loadErr
		},
	)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("当前番剧没有解析到可播放剧集")
	}
	return groups, nil
}

func loadAggregatedEpisodeGroups(application *app.App, anime source.AggregatedAnime) ([]source.EpisodeGroup, error) {
	if len(anime.Hits) == 0 {
		return nil, nil
	}

	type episodeFetchMessage struct {
		sourceName string
		episodes   []source.Episode
		err        error
	}

	resultCh := make(chan episodeFetchMessage, len(anime.Hits))
	for _, hit := range anime.Hits {
		go func(hit source.AnimeHit) {
			src := application.GetSourceByName(hit.SourceName)
			if src == nil {
				resultCh <- episodeFetchMessage{
					sourceName: hit.SourceName,
					err:        fmt.Errorf("未找到媒体源"),
				}
				return
			}

			episodes, err := src.GetEpisodes(hit.Anime.URL)
			resultCh <- episodeFetchMessage{
				sourceName: hit.SourceName,
				episodes:   episodes,
				err:        err,
			}
		}(hit)
	}

	aggregatedEpisodes := make([]source.Episode, 0)
	var failures []string
	for range anime.Hits {
		result := <-resultCh
		if result.err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.sourceName, result.err))
			continue
		}
		aggregatedEpisodes = append(aggregatedEpisodes, withEpisodeSourceName(result.sourceName, result.episodes)...)
	}

	groups := source.GroupEpisodes(aggregatedEpisodes)
	if len(groups) == 0 && len(failures) > 0 {
		return nil, fmt.Errorf("所有片源都获取剧集失败了:\n%s", strings.Join(failures, "\n"))
	}
	return groups, nil
}

func withEpisodeSourceName(sourceName string, episodes []source.Episode) []source.Episode {
	tagged := make([]source.Episode, len(episodes))
	for i, episode := range episodes {
		tagged[i] = episode
		tagged[i].SourceName = sourceName
	}
	return tagged
}

func clampIndex(index, total int) int {
	if total <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= total {
		return total - 1
	}
	return index
}
