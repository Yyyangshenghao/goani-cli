package workflow

import (
	"errors"
	"testing"

	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

func TestResolvePlaybackNavigation(t *testing.T) {
	tests := []struct {
		name             string
		action           tui.PlaybackAction
		currentEpisode   int
		totalEpisodes    int
		wantMode         playbackNavigationMode
		wantEpisodeIndex int
	}{
		{
			name:             "previous episode moves backward",
			action:           tui.PlaybackActionPreviousEpisode,
			currentEpisode:   3,
			totalEpisodes:    8,
			wantMode:         playbackNavigationPlayEpisode,
			wantEpisodeIndex: 2,
		},
		{
			name:             "next episode moves forward",
			action:           tui.PlaybackActionNextEpisode,
			currentEpisode:   3,
			totalEpisodes:    8,
			wantMode:         playbackNavigationPlayEpisode,
			wantEpisodeIndex: 4,
		},
		{
			name:             "previous episode at beginning stays on line selection",
			action:           tui.PlaybackActionPreviousEpisode,
			currentEpisode:   0,
			totalEpisodes:    8,
			wantMode:         playbackNavigationStayOnLineSelection,
			wantEpisodeIndex: 0,
		},
		{
			name:             "episode list returns to selector",
			action:           tui.PlaybackActionEpisodeList,
			currentEpisode:   3,
			totalEpisodes:    8,
			wantMode:         playbackNavigationReturnEpisodeList,
			wantEpisodeIndex: 3,
		},
		{
			name:             "anime list returns to anime selector",
			action:           tui.PlaybackActionAnimeList,
			currentEpisode:   3,
			totalEpisodes:    8,
			wantMode:         playbackNavigationReturnAnimeList,
			wantEpisodeIndex: 3,
		},
		{
			name:             "home exits flow",
			action:           tui.PlaybackActionHome,
			currentEpisode:   3,
			totalEpisodes:    8,
			wantMode:         playbackNavigationExitFlow,
			wantEpisodeIndex: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolvePlaybackNavigation(tt.action, tt.currentEpisode, tt.totalEpisodes)

			if got.mode != tt.wantMode {
				t.Fatalf("expected mode %v, got %v", tt.wantMode, got.mode)
			}
			if got.episodeIndex != tt.wantEpisodeIndex {
				t.Fatalf("expected episode index %d, got %d", tt.wantEpisodeIndex, got.episodeIndex)
			}
		})
	}
}

func TestEpisodeCandidateLabel(t *testing.T) {
	group := source.EpisodeGroup{Name: "第1集"}

	tests := []struct {
		name      string
		index     int
		candidate source.EpisodeCandidate
		want      string
	}{
		{
			name:  "falls back to numbered label when source missing",
			index: 0,
			candidate: source.EpisodeCandidate{
				Name: "第1集",
			},
			want: "线路1",
		},
		{
			name:  "includes source for default line label",
			index: 1,
			candidate: source.EpisodeCandidate{
				Name:       "第1集",
				SourceName: "源A",
			},
			want: "源A / 线路2",
		},
		{
			name:  "prefixes source for custom label",
			index: 0,
			candidate: source.EpisodeCandidate{
				Name:       "备用线路",
				SourceName: "源B",
			},
			want: "源B / 备用线路",
		},
		{
			name:  "avoids repeating source name twice",
			index: 0,
			candidate: source.EpisodeCandidate{
				Name:       "源C / 高清线路",
				SourceName: "源C",
			},
			want: "源C / 高清线路",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := episodeCandidateLabel(group, tt.index, tt.candidate)
			if got != tt.want {
				t.Fatalf("unexpected label: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestPlayWithRequestContextPrefersFFmpegBridgeForM3U8(t *testing.T) {
	oldFFmpeg := startDetachedFFmpegHLSBridge
	oldProxy := startDetachedHLSProxy
	t.Cleanup(func() {
		startDetachedFFmpegHLSBridge = oldFFmpeg
		startDetachedHLSProxy = oldProxy
	})

	startDetachedFFmpegHLSBridge = func(ctx player.StreamRequestContext) (string, error) {
		if ctx.SourceURL != "https://media.example.com/master.m3u8" {
			t.Fatalf("unexpected source url: got %q", ctx.SourceURL)
		}
		return "http://127.0.0.1:9000/live.ts", nil
	}
	startDetachedHLSProxy = func(ctx player.StreamRequestContext) (string, error) {
		t.Fatalf("proxy should not be used when ffmpeg bridge succeeds")
		return "", nil
	}

	p := &fakePlaybackPlayer{}
	err := playWithRequestContext(p, player.StreamRequestContext{
		SourceURL: "https://media.example.com/master.m3u8",
	})
	if err != nil {
		t.Fatalf("playWithRequestContext returned error: %v", err)
	}
	if len(p.played) != 1 || p.played[0] != "http://127.0.0.1:9000/live.ts" {
		t.Fatalf("unexpected played urls: %v", p.played)
	}
}

func TestPlayWithRequestContextFallsBackToProxyWhenFFmpegFails(t *testing.T) {
	oldFFmpeg := startDetachedFFmpegHLSBridge
	oldProxy := startDetachedHLSProxy
	t.Cleanup(func() {
		startDetachedFFmpegHLSBridge = oldFFmpeg
		startDetachedHLSProxy = oldProxy
	})

	startDetachedFFmpegHLSBridge = func(ctx player.StreamRequestContext) (string, error) {
		return "", errors.New("ffmpeg unavailable")
	}
	startDetachedHLSProxy = func(ctx player.StreamRequestContext) (string, error) {
		return "http://127.0.0.1:9001/master.m3u8", nil
	}

	p := &fakePlaybackPlayer{}
	err := playWithRequestContext(p, player.StreamRequestContext{
		SourceURL: "https://media.example.com/master.m3u8",
	})
	if err != nil {
		t.Fatalf("playWithRequestContext returned error: %v", err)
	}
	if len(p.played) != 1 || p.played[0] != "http://127.0.0.1:9001/master.m3u8" {
		t.Fatalf("unexpected played urls: %v", p.played)
	}
}

type fakePlaybackPlayer struct {
	played []string
}

func (p *fakePlaybackPlayer) Name() string {
	return "fake"
}

func (p *fakePlaybackPlayer) Play(url string) error {
	p.played = append(p.played, url)
	return nil
}

func (p *fakePlaybackPlayer) PlayWithArgs(url string, args []string) error {
	p.played = append(p.played, url)
	return nil
}
