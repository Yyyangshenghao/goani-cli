package workflow

import (
	"fmt"
	"os"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
	consoleui "github.com/Yyyangshenghao/goani-cli/internal/ui/console"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

const playbackUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type playbackLaunchResult struct {
	playerName   string
	episodeLabel string
	lineLabel    string
}

type playbackNavigationMode int

const (
	playbackNavigationStayOnLineSelection playbackNavigationMode = iota
	playbackNavigationPlayEpisode
	playbackNavigationReturnEpisodeList
	playbackNavigationReturnAnimeList
	playbackNavigationExitFlow
)

type playbackNavigationResult struct {
	mode         playbackNavigationMode
	episodeIndex int
}

// ShowAnimeListAndSelect 承接经典 CLI 的搜索后半段：选番、选集、播放。
func ShowAnimeListAndSelect(application *app.App, animes []source.Anime, sourceName string) {
	if len(animes) == 0 {
		consoleui.Info("未找到结果")
		return
	}

	consoleui.Success("[%s] 找到 %d 条结果", sourceName, len(animes))
	for i, a := range animes {
		fmt.Printf("  %d. %s\n", i+1, a.Name)
	}

	idx, err := consoleui.Select("选择动漫", len(animes), func(i int) string { return animes[i].Name })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	src := application.GetSourceByName(sourceName)
	if src == nil {
		consoleui.Error("无法获取媒体源")
		os.Exit(1)
	}

	showEpisodesAndSelectWithSource(application, src, animes[idx].URL)
}

// PlayEpisodeGroupCLI 让经典 CLI 和脚本路径也共享统一的多线路播放逻辑。
func PlayEpisodeGroupCLI(application *app.App, src *webselector.WebSelectorSource, group source.EpisodeGroup) error {
	p, err := application.GetPlayer()
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("未找到可用播放器\n请先配置: goani config player <name> <path>")
	}

	var attempts []string
	for i, candidate := range group.Candidates {
		videoURL, err := src.GetVideoURL(candidate.URL)
		label := episodeCandidateLabel(group, i, candidate)
		if err != nil {
			attempts = append(attempts, fmt.Sprintf("%s 解析失败: %v", label, err))
			continue
		}
		consoleui.Info("使用 %s 播放 %s...", p.Name(), label)
		if err := playWithRequestContext(p, videoURL, candidate.URL); err != nil {
			attempts = append(attempts, fmt.Sprintf("%s 播放失败: %v", label, err))
			continue
		}
		consoleui.Success("播放器已启动")
		return nil
	}

	return fmt.Errorf("这一集的所有线路都失败了:\n%s", strings.Join(attempts, "\n"))
}

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

func showEpisodesAndSelectWithSource(application *app.App, src *webselector.WebSelectorSource, animeURL string) {
	episodes, err := src.GetEpisodes(animeURL)
	if err != nil {
		consoleui.Error("获取剧集失败: %v", err)
		os.Exit(1)
	}
	groups := source.GroupEpisodes(episodes)
	if len(groups) == 0 {
		consoleui.Info("没有可用剧集")
		return
	}

	consoleui.Success("找到 %d 集", len(groups))
	printEpisodeGroups(groups, 20)

	epIdx, err := consoleui.Select("选择剧集", len(groups), func(i int) string { return groups[i].Label() })
	if err != nil {
		fmt.Println("已取消")
		return
	}

	if err := PlayEpisodeGroupCLI(application, src, groups[epIdx]); err != nil {
		consoleui.Error("%v", err)
		os.Exit(1)
	}
}

func printEpisodeGroups(groups []source.EpisodeGroup, max int) {
	for i, ep := range groups {
		if i >= max {
			fmt.Printf("  ... 还有 %d 集\n", len(groups)-max)
			break
		}
		fmt.Printf("  %d. %s\n", i+1, ep.Label())
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
	if err := playWithRequestContext(p, candidate.videoURL, candidate.episodeURL); err != nil {
		return nil, fmt.Errorf("%s 播放失败: %w", candidate.name, err)
	}

	return &playbackLaunchResult{
		playerName:   p.Name(),
		episodeLabel: group.Label(),
		lineLabel:    candidate.name,
	}, nil
}

func episodeCandidateLabel(group source.EpisodeGroup, index int, candidate source.EpisodeCandidate) string {
	name := strings.TrimSpace(candidate.Name)
	sourceName := strings.TrimSpace(candidate.SourceName)

	switch {
	case name == "" || name == group.Name:
		if sourceName == "" {
			return fmt.Sprintf("线路%d", index+1)
		}
		return fmt.Sprintf("%s / 线路%d", sourceName, index+1)
	case sourceName == "":
		return name
	case strings.Contains(name, sourceName):
		return name
	default:
		return fmt.Sprintf("%s / %s", sourceName, name)
	}
}

func showTUIErrorScreen(title, message string) {
	if err := tui.RunTextTUI(title, message); err != nil {
		consoleui.Error("%s: %s", title, message)
	}
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

func resolvePlaybackNavigation(action tui.PlaybackAction, currentEpisodeIndex, totalEpisodes int) playbackNavigationResult {
	result := playbackNavigationResult{
		mode:         playbackNavigationStayOnLineSelection,
		episodeIndex: currentEpisodeIndex,
	}

	switch action {
	case tui.PlaybackActionPreviousEpisode:
		if currentEpisodeIndex > 0 {
			result.mode = playbackNavigationPlayEpisode
			result.episodeIndex = currentEpisodeIndex - 1
		}
	case tui.PlaybackActionNextEpisode:
		if currentEpisodeIndex+1 < totalEpisodes {
			result.mode = playbackNavigationPlayEpisode
			result.episodeIndex = currentEpisodeIndex + 1
		}
	case tui.PlaybackActionEpisodeList:
		result.mode = playbackNavigationReturnEpisodeList
	case tui.PlaybackActionAnimeList:
		result.mode = playbackNavigationReturnAnimeList
	case tui.PlaybackActionHome, tui.PlaybackActionQuit:
		result.mode = playbackNavigationExitFlow
	}

	return result
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

// playWithRequestContext 会根据播放器和线路格式决定是否启用兼容层。
// 目前主要是 PotPlayer 播放受保护的 m3u8 时，通过本地 HLS 代理兜底。
func playWithRequestContext(p interface {
	Name() string
	Play(string) error
	PlayWithArgs(string, []string) error
}, videoURL, referer string) error {
	if p.Name() == "potplayer" && detectMediaFormat(videoURL) == "m3u8" {
		proxiedURL, err := player.StartDetachedHLSProxy(videoURL, referer, playbackUserAgent)
		if err != nil {
			return fmt.Errorf("启动本地 HLS 代理失败: %w", err)
		}
		return p.Play(proxiedURL)
	}

	return p.Play(videoURL)
}
