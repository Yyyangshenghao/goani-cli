package workflow

import (
	"fmt"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	consoleui "github.com/Yyyangshenghao/goani-cli/internal/ui/console"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

const playbackUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var startDetachedFFmpegHLSBridge = player.StartDetachedFFmpegHLSBridge
var startDetachedHLSProxy = player.StartDetachedHLSProxy

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

// playWithRequestContext 会根据线路格式决定是否启用兼容层。
// 很多站点的 m3u8/分片请求依赖 Referer、User-Agent、Cookie 或额外请求头，
// 这里统一优先走 goani 自己的本地 HLS 代理，再回退到播放器原生 / ffmpeg。
func playWithRequestContext(p interface {
	Name() string
	Play(string) error
	PlayWithArgs(string, []string) error
}, requestContext player.StreamRequestContext) error {
	requestContext = requestContext.Normalized()
	if detectMediaFormat(requestContext.SourceURL) == "m3u8" {
		proxiedURL, proxyErr := startDetachedHLSProxy(requestContext)
		if proxyErr == nil {
			return p.Play(proxiedURL)
		}

		var nativeErr error
		if p.Name() == "mpv" {
			if err := p.PlayWithArgs(requestContext.SourceURL, requestContext.MPVHTTPArgs()); err == nil {
				return nil
			} else {
				nativeErr = err
			}
		}

		streamURL, ffmpegErr := startDetachedFFmpegHLSBridge(requestContext)
		if ffmpegErr == nil {
			return p.Play(streamURL)
		}
		if nativeErr != nil {
			return fmt.Errorf("启动本地 HLS 代理失败: %v；mpv 原生播放失败: %v；启动 ffmpeg 桥接失败: %w", proxyErr, nativeErr, ffmpegErr)
		}
		return fmt.Errorf("启动本地 HLS 代理失败: %v；启动 ffmpeg 桥接失败: %w", proxyErr, ffmpegErr)
	}

	return p.Play(requestContext.SourceURL)
}

func buildPlaybackRequestContext(application *app.App, sourceName, videoURL, referer string) player.StreamRequestContext {
	ctx := player.StreamRequestContext{
		SourceURL: videoURL,
		Referer:   referer,
		UserAgent: playbackUserAgent,
	}

	sourceName = strings.TrimSpace(sourceName)
	if sourceName == "" {
		return ctx
	}

	rawSource := application.SourceManager.GetByName(sourceName)
	if rawSource == nil {
		return ctx
	}

	matchVideo := rawSource.Arguments.SearchConfig.MatchVideo
	if strings.TrimSpace(matchVideo.Cookies) != "" {
		ctx.Cookies = strings.TrimSpace(matchVideo.Cookies)
	}
	if len(matchVideo.AddHeadersToVideo) > 0 {
		ctx.Headers = cloneStringMap(matchVideo.AddHeadersToVideo)
	}
	return ctx
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
