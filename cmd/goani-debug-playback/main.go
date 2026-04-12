package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/player"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

const playbackSampleUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

const playbackSampleProbeTimeout = 15 * time.Second

var defaultPlaybackSampleKeywords = []string{
	"葬送的芙莉莲",
	"鬼灭之刃",
	"进击的巨人",
	"海贼王",
	"名侦探柯南",
	"孤独摇滚",
	"药屋少女的呢喃",
	"咒术回战",
}

func main() {
	output := flag.String("output", filepath.Join("internal", "player", "testdata", "m3u8_samples.local.json"), "样本输出路径")
	flag.Parse()

	keywords := defaultPlaybackSampleKeywords
	if args := flag.Args(); len(args) > 0 {
		keywords = args
	}

	fmt.Println("=== goani m3u8 样本采集 ===")
	fmt.Printf("输出文件: %s\n", *output)
	fmt.Printf("关键词数: %d\n\n", len(keywords))

	application := app.New()
	loadedSources := application.SourceManager.GetEnabled()
	if len(loadedSources) == 0 {
		fmt.Println("未找到启用中的媒体源，请先开启至少一个源")
		os.Exit(1)
	}

	samples := make([]player.PlaybackSample, 0, len(keywords))
	for _, keyword := range keywords {
		fmt.Printf("【采集】%s\n", keyword)
		sample, ok := collectSampleForKeyword(loadedSources, keyword)
		if !ok {
			fmt.Println("  未找到可用 m3u8 样本")
			fmt.Println()
			continue
		}

		samples = append(samples, sample)
		fmt.Printf("  源: %s\n", sample.SourceName)
		fmt.Printf("  动漫: %s\n", sample.AnimeName)
		fmt.Printf("  剧集: %s\n", sample.EpisodeName)
		fmt.Printf("  直链: %s\n\n", sample.VideoURL)
	}

	if len(samples) == 0 {
		fmt.Println("没有采集到任何 m3u8 样本")
		os.Exit(1)
	}

	file := player.PlaybackSampleFile{
		CollectedAt: time.Now().Format("2006-01-02 15:04:05"),
		Generator:   "goani-debug-playback",
		Cases:       samples,
	}
	if err := player.SavePlaybackSampleFile(*output, file); err != nil {
		fmt.Printf("保存样本失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("完成: 共保存 %d 条 m3u8 样本\n", len(samples))
}

func collectSampleForKeyword(loadedSources []source.MediaSource, keyword string) (player.PlaybackSample, bool) {
	for _, mediaSource := range loadedSources {
		sample, ok := collectSampleFromSource(mediaSource, keyword)
		if ok {
			return sample, true
		}
	}
	return player.PlaybackSample{}, false
}

func collectSampleFromSource(mediaSource source.MediaSource, keyword string) (player.PlaybackSample, bool) {
	ws := webselector.New(mediaSource).WithTimeout(12 * time.Second)
	results, err := ws.Search(keyword)
	if err != nil || len(results) == 0 {
		return player.PlaybackSample{}, false
	}

	animeLimit := minInt(len(results), 3)
	for _, anime := range results[:animeLimit] {
		episodes, err := ws.GetEpisodes(anime.URL)
		if err != nil || len(episodes) == 0 {
			continue
		}

		episodeLimit := minInt(len(episodes), 5)
		for _, episode := range episodes[:episodeLimit] {
			videoURL, err := ws.GetVideoURL(episode.URL)
			if err != nil || !strings.Contains(strings.ToLower(videoURL), ".m3u8") {
				continue
			}

			matchVideo := mediaSource.Arguments.SearchConfig.MatchVideo
			sample := player.PlaybackSample{
				CollectedAt: time.Now().Format("2006-01-02 15:04:05"),
				Keyword:     keyword,
				AnimeName:   strings.TrimSpace(anime.Name),
				EpisodeName: strings.TrimSpace(episode.Name),
				SourceID:    mediaSource.PreferenceID(),
				SourceName:  strings.TrimSpace(mediaSource.Arguments.Name),
				EpisodeURL:  episode.URL,
				VideoURL:    videoURL,
				Cookies:     strings.TrimSpace(matchVideo.Cookies),
				Headers:     normalizedHeaders(matchVideo.AddHeadersToVideo),
			}
			if !sampleLooksPlayable(sample) {
				continue
			}
			return sample, true
		}
	}

	return player.PlaybackSample{}, false
}

func normalizedHeaders(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}

	headers := make(map[string]string, len(input))
	for key, value := range input {
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			continue
		}
		headers[trimmedKey] = trimmedValue
	}
	if len(headers) == 0 {
		return nil
	}
	return headers
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sampleLooksPlayable(sample player.PlaybackSample) bool {
	req, err := http.NewRequest(http.MethodGet, sample.VideoURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", playbackSampleUserAgent)
	if strings.TrimSpace(sample.EpisodeURL) != "" {
		req.Header.Set("Referer", strings.TrimSpace(sample.EpisodeURL))
	}
	if strings.TrimSpace(sample.Cookies) != "" {
		req.Header.Set("Cookie", strings.TrimSpace(sample.Cookies))
	}
	for key, value := range sample.Headers {
		req.Header.Set(strings.TrimSpace(key), strings.TrimSpace(value))
	}

	client := &http.Client{Timeout: playbackSampleProbeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil {
		return false
	}
	return len(body) > 0
}
