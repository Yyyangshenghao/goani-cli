package player

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	manualPlaybackEnv          = "GOANI_RUN_MANUAL_PLAYER_TESTS"
	livePlaybackSmokeEnv       = "GOANI_RUN_LIVE_PLAYBACK_SMOKE"
	playbackSamplePathEnv      = "GOANI_PLAYBACK_SAMPLE_FILE"
	playbackPlayerFilterEnv    = "GOANI_PLAYBACK_PLAYERS"
	defaultPlaybackSamplePath  = "testdata/m3u8_samples.local.json"
	testPlaybackUserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	livePlaybackHTTPTimeout    = 20 * time.Second
	livePlaybackReadBytesLimit = 2048
)

func TestLivePlaybackSmoke(t *testing.T) {
	if os.Getenv(livePlaybackSmokeEnv) != "1" {
		t.Skipf("设置 %s=1 后再运行真实 m3u8 smoke 测试", livePlaybackSmokeEnv)
	}

	samples := mustLoadPlaybackSamples(t)
	client := &http.Client{Timeout: livePlaybackHTTPTimeout}

	for _, sample := range samples.Cases {
		sample := sample
		t.Run(sampleDisplayName(sample), func(t *testing.T) {
			localURL, err := startPlaybackURLForSample(sample)
			if err != nil {
				t.Fatalf("准备本地播放地址失败: %v", err)
			}

			resp, err := client.Get(localURL)
			if err != nil {
				t.Fatalf("请求本地播放地址失败: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status code: got %d want %d", resp.StatusCode, http.StatusOK)
			}

			body, err := io.ReadAll(io.LimitReader(resp.Body, livePlaybackReadBytesLimit))
			if err != nil {
				t.Fatalf("读取播放数据失败: %v", err)
			}
			if len(body) == 0 {
				t.Fatalf("未读取到任何播放数据")
			}
		})
	}
}

func TestManualPlayerPlaybackMatrix(t *testing.T) {
	if os.Getenv(manualPlaybackEnv) != "1" {
		t.Skipf("设置 %s=1 后再运行手工播放器矩阵测试", manualPlaybackEnv)
	}

	samples := mustLoadPlaybackSamples(t)
	players := availablePlayersForManualTest(t)
	if len(players) == 0 {
		t.Fatalf("没有找到可用播放器，请先在配置里设置播放器路径")
	}

	reader := bufio.NewReader(os.Stdin)
	for _, playbackPlayer := range players {
		playbackPlayer := playbackPlayer
		t.Run(playbackPlayer.Name(), func(t *testing.T) {
			for _, sample := range samples.Cases {
				sample := sample
				t.Run(sampleDisplayName(sample), func(t *testing.T) {
					localURL, err := startPlaybackURLForSample(sample)
					if err != nil {
						t.Fatalf("准备本地播放地址失败: %v", err)
					}

					fmt.Printf("\n[%s] 正在播放: %s\n本地地址: %s\n", playbackPlayer.Name(), sampleDisplayName(sample), localURL)
					fmt.Println("请确认播放器已经正常播放。")
					fmt.Println("确认成功后按回车继续；输入 n 再回车记为失败；输入 s 再回车记为跳过。")

					if err := playbackPlayer.Play(localURL); err != nil {
						t.Fatalf("启动播放器失败: %v", err)
					}

					answer, err := reader.ReadString('\n')
					if err != nil {
						t.Fatalf("读取人工确认结果失败: %v", err)
					}

					switch strings.ToLower(strings.TrimSpace(answer)) {
					case "", "y", "yes":
						return
					case "s", "skip":
						t.Skip("人工确认跳过")
					default:
						t.Fatalf("人工确认播放失败")
					}
				})
			}
		})
	}
}

func mustLoadPlaybackSamples(t *testing.T) *PlaybackSampleFile {
	t.Helper()

	path := os.Getenv(playbackSamplePathEnv)
	if strings.TrimSpace(path) == "" {
		path = defaultPlaybackSamplePath
	}

	file, err := LoadPlaybackSampleFile(path)
	if err != nil {
		t.Fatalf("读取样本文件失败: %v\n请先运行: go run .\\cmd\\goani-debug-playback", err)
	}
	if len(file.Cases) == 0 {
		t.Fatalf("样本文件为空: %s", path)
	}
	return file
}

func availablePlayersForManualTest(t *testing.T) []Player {
	t.Helper()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("读取播放器配置失败: %v", err)
	}

	players := configuredPlayersForManualTest(cfg)
	if filter := parsePlayerFilter(os.Getenv(playbackPlayerFilterEnv)); len(filter) > 0 {
		filtered := make([]Player, 0, len(players))
		for _, playbackPlayer := range players {
			if _, ok := filter[playbackPlayer.Name()]; ok {
				filtered = append(filtered, playbackPlayer)
			}
		}
		return filtered
	}
	return players
}

func configuredPlayersForManualTest(cfg *Config) []Player {
	if cfg == nil {
		return nil
	}

	players := make([]Player, 0, len(SupportedPlayers()))
	for _, name := range orderedPlayers(cfg.DefaultPlayer) {
		path := strings.TrimSpace(cfg.GetPath(name))
		if path == "" {
			continue
		}

		playbackPlayer := newPlayer(name, path)
		if playbackPlayer == nil || !playbackPlayer.IsAvailable() {
			continue
		}
		players = append(players, playbackPlayer)
	}
	return players
}

func parsePlayerFilter(raw string) map[string]struct{} {
	parts := strings.Split(raw, ",")
	filter := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(strings.ToLower(part))
		if name == "" {
			continue
		}
		filter[name] = struct{}{}
	}
	if len(filter) == 0 {
		return nil
	}
	return filter
}

func startPlaybackURLForSample(sample PlaybackSample) (string, error) {
	requestContext := sample.StreamRequestContext(testPlaybackUserAgent)
	if _, err := ffmpegLookPath("ffmpeg"); err == nil {
		localURL, err := startInProcessPlaybackServer(func(out io.Writer) error {
			return ServeFFmpegHLSBridge(requestContext, out)
		})
		if err == nil {
			return localURL, nil
		}
	}

	return startInProcessPlaybackServer(func(out io.Writer) error {
		return ServeHLSProxy(requestContext, out)
	})
}

func sampleDisplayName(sample PlaybackSample) string {
	parts := []string{
		strings.TrimSpace(sample.Keyword),
		strings.TrimSpace(sample.AnimeName),
		strings.TrimSpace(sample.EpisodeName),
		strings.TrimSpace(sample.SourceName),
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	if len(filtered) == 0 {
		return "unknown-sample"
	}
	return strings.Join(filtered, " / ")
}

func startInProcessPlaybackServer(run func(io.Writer) error) (string, error) {
	reader, writer := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		errCh <- run(writer)
	}()

	urlCh := make(chan string, 1)
	readErrCh := make(chan error, 1)
	go func() {
		buffer := bufio.NewReader(reader)
		line, err := buffer.ReadString('\n')
		if err != nil {
			readErrCh <- err
			return
		}
		urlCh <- strings.TrimSpace(line)
	}()

	select {
	case localURL := <-urlCh:
		if strings.TrimSpace(localURL) == "" {
			return "", fmt.Errorf("播放服务没有返回本地地址")
		}
		return localURL, nil
	case err := <-readErrCh:
		return "", fmt.Errorf("读取播放服务地址失败: %w", err)
	case err := <-errCh:
		return "", fmt.Errorf("启动播放服务失败: %w", err)
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("启动播放服务超时")
	}
}
