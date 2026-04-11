package workflow

import (
	"fmt"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
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
