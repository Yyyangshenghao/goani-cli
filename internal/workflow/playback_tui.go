package workflow

import (
	"fmt"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

func showAggregatedSelectionFlow(application *app.App, animes []source.AggregatedAnime, selectedAnimeIndex int) error {
	if len(animes) == 0 {
		return nil
	}

	selectedAnimeIndex = clampIndex(selectedAnimeIndex, len(animes))

animeLoop:
	for {
		anime := animes[selectedAnimeIndex]
		groups, err := loadAggregatedEpisodeGroupsTUI(application, anime)
		if err != nil {
			showTUIErrorScreen("聚合剧集失败", err.Error())
			nextAnime, nextIndex, err := tui.RunAggregatedAnimeSelectionTUI(animes, selectedAnimeIndex)
			if err != nil {
				return fmt.Errorf("打开聚合番剧列表失败: %w", err)
			}
			if nextAnime == nil {
				return nil
			}
			selectedAnimeIndex = nextIndex
			continue
		}

		selectedEpisodeIndex := 0

	episodeLoop:
		for {
			episode, episodeIndex, err := tui.RunEpisodeSelectionTUIWithSelection(anime.Name, groups, selectedEpisodeIndex)
			if err != nil {
				return fmt.Errorf("打开选集界面失败: %w", err)
			}
			if episode == nil {
				nextAnime, nextIndex, err := tui.RunAggregatedAnimeSelectionTUI(animes, selectedAnimeIndex)
				if err != nil {
					return fmt.Errorf("打开聚合番剧列表失败: %w", err)
				}
				if nextAnime == nil {
					return nil
				}
				selectedAnimeIndex = nextIndex
				continue animeLoop
			}
			selectedEpisodeIndex = episodeIndex

		playCurrentEpisode:
			for {
				currentEpisode := groups[selectedEpisodeIndex]

				lines, err := loadResolvedEpisodeLinesTUI(application, currentEpisode)
				if err != nil {
					showTUIErrorScreen("线路解析失败", err.Error())
					continue episodeLoop
				}

			lineLoop:
				for {
					lineResult, err := tui.RunLineSelectionTUI(anime.Name, currentEpisode.Label(), buildLineSelectionItems(lines))
					if err != nil {
						return fmt.Errorf("打开线路选择页失败: %w", err)
					}
					if lineResult == nil {
						continue episodeLoop
					}

					line := lines[lineResult.Index]
					launch, err := playResolvedEpisodeCandidateTUI(application, currentEpisode, line)
					if err != nil {
						showTUIErrorScreen("播放失败", err.Error())
						continue lineLoop
					}

					nextAction, err := tui.RunPlaybackPageTUI(
						anime.Name,
						launch.episodeLabel,
						launch.playerName,
						launch.lineLabel,
						selectedEpisodeIndex > 0,
						selectedEpisodeIndex < len(groups)-1,
					)
					if err != nil {
						return fmt.Errorf("打开播放状态页失败: %w", err)
					}

					navigation := resolvePlaybackNavigation(nextAction, selectedEpisodeIndex, len(groups))
					selectedEpisodeIndex = navigation.episodeIndex

					switch navigation.mode {
					case playbackNavigationPlayEpisode:
						continue playCurrentEpisode
					case playbackNavigationReturnEpisodeList:
						continue episodeLoop
					case playbackNavigationReturnAnimeList:
						nextAnime, nextIndex, err := tui.RunAggregatedAnimeSelectionTUI(animes, selectedAnimeIndex)
						if err != nil {
							return fmt.Errorf("打开聚合番剧列表失败: %w", err)
						}
						if nextAnime == nil {
							return nil
						}
						selectedAnimeIndex = nextIndex
						continue animeLoop
					case playbackNavigationExitFlow:
						return nil
					default:
						continue lineLoop
					}
				}
			}
		}
	}
}

func playResolvedEpisodeCandidateTUI(application *app.App, group source.EpisodeGroup, candidate resolvedEpisodeCandidate) (*playbackLaunchResult, error) {
	if candidate.err != nil {
		return nil, candidate.err
	}

	p, err := application.GetPlayer()
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("未找到可用播放器\n\n请先配置: goani config player <name> <path>")
	}
	if strings.TrimSpace(candidate.videoURL) == "" {
		return nil, fmt.Errorf("当前线路没有可播放的视频链接")
	}
	requestContext := buildPlaybackRequestContext(application, candidate.sourceName, candidate.videoURL, candidate.episodeURL)
	if err := playWithRequestContext(p, requestContext); err != nil {
		return nil, fmt.Errorf("%s 播放失败: %w", candidate.name, err)
	}

	return &playbackLaunchResult{
		playerName:   p.Name(),
		episodeLabel: group.Label(),
		lineLabel:    candidate.name,
	}, nil
}

func loadResolvedEpisodeLinesTUI(application *app.App, group source.EpisodeGroup) ([]resolvedEpisodeCandidate, error) {
	var lines []resolvedEpisodeCandidate
	err := tui.RunLoadingTUI(
		"正在解析线路",
		"正在获取视频链接和清晰度，最长等待 5 秒",
		func() error {
			lines = resolveEpisodeCandidates(application, group)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("当前剧集没有解析到可用线路")
	}
	return lines, nil
}
