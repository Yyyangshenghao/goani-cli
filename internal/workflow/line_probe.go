package workflow

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	tui "github.com/Yyyangshenghao/goani-cli/internal/ui/tui"
)

const (
	lineResolutionTimeout   = 5 * time.Second
	playlistProbeTimeout    = 1500 * time.Millisecond
	lineResolutionErrorText = "解析超时（超过 5 秒）"
)

// resolvedEpisodeCandidate 表示某一条候选线路的解析结果。
// 它保留了原始剧集页地址、最终视频地址，以及尽力探测出的格式和清晰度。
type resolvedEpisodeCandidate struct {
	name       string
	episodeURL string
	videoURL   string
	format     string
	quality    string
	err        error
}

type resolvedCandidateMessage struct {
	index int
	item  resolvedEpisodeCandidate
}

// resolveEpisodeCandidates 会并发解析当前剧集下的所有线路，并把总等待时间限制在 5 秒以内。
// 超时的线路会保留在列表里，但会被标记成失败状态，避免整页被慢线路拖住。
func resolveEpisodeCandidates(application *app.App, group source.EpisodeGroup) []resolvedEpisodeCandidate {
	resolved := make([]resolvedEpisodeCandidate, len(group.Candidates))
	resultCh := make(chan resolvedCandidateMessage, len(group.Candidates))

	for i, candidate := range group.Candidates {
		label := episodeCandidateLabel(group, i, candidate)
		resolved[i] = resolvedEpisodeCandidate{
			name:       label,
			episodeURL: candidate.URL,
		}

		go func(index int, name string, candidate source.EpisodeCandidate) {
			resultCh <- resolvedCandidateMessage{
				index: index,
				item:  resolveEpisodeCandidate(application, name, candidate),
			}
		}(i, label, candidate)
	}

	timer := time.NewTimer(lineResolutionTimeout)
	defer timer.Stop()

	received := 0
	for received < len(group.Candidates) {
		select {
		case msg := <-resultCh:
			resolved[msg.index] = msg.item
			received++
		case <-timer.C:
			for i := range resolved {
				if resolved[i].err == nil && strings.TrimSpace(resolved[i].videoURL) == "" {
					resolved[i].err = fmt.Errorf(lineResolutionErrorText)
					resolved[i].quality = "未知"
					resolved[i].format = "未知"
				}
			}
			return resolved
		}
	}

	return resolved
}

func resolveEpisodeCandidate(application *app.App, name string, candidate source.EpisodeCandidate) resolvedEpisodeCandidate {
	item := resolvedEpisodeCandidate{
		name:       name,
		episodeURL: candidate.URL,
	}

	sourceName := strings.TrimSpace(candidate.SourceName)
	if sourceName == "" {
		item.err = fmt.Errorf("当前线路缺少片源信息")
		return item
	}

	src := application.GetSourceByName(sourceName)
	if src == nil {
		item.err = fmt.Errorf("未找到片源: %s", sourceName)
		return item
	}

	resolver := src.WithTimeout(lineResolutionTimeout)
	videoURL, err := resolver.GetVideoURL(candidate.URL)
	if err != nil {
		item.err = fmt.Errorf("解析直链失败: %w", err)
		return item
	}

	item.videoURL = videoURL
	item.format = detectMediaFormat(videoURL)
	item.quality = detectQualityFromURL(videoURL)

	if item.format == "m3u8" {
		probeClient := resolver.WithTimeout(playlistProbeTimeout).Client()
		if playlistQuality, err := probeM3U8Quality(probeClient, videoURL); err == nil && strings.TrimSpace(playlistQuality) != "" {
			item.quality = playlistQuality
		}
	}
	if item.quality == "" {
		item.quality = "未知"
	}

	return item
}

func buildLineSelectionItems(items []resolvedEpisodeCandidate) []tui.LineSelectionItem {
	result := make([]tui.LineSelectionItem, 0, len(items))
	for _, item := range items {
		entry := tui.LineSelectionItem{
			Title:      item.name,
			EpisodeURL: item.episodeURL,
			VideoURL:   item.videoURL,
			Format:     item.format,
			Quality:    item.quality,
		}
		if item.err != nil {
			entry.Error = item.err.Error()
		}
		result = append(result, entry)
	}
	return result
}

func detectMediaFormat(rawURL string) string {
	lower := strings.ToLower(rawURL)
	switch {
	case strings.Contains(lower, ".m3u8"):
		return "m3u8"
	case strings.Contains(lower, ".mp4"):
		return "mp4"
	case strings.Contains(lower, ".flv"):
		return "flv"
	case strings.Contains(lower, ".mkv"):
		return "mkv"
	default:
		return "未知"
	}
}

func detectQualityFromURL(rawURL string) string {
	lower := strings.ToLower(rawURL)
	patterns := []string{"2160", "1440", "1080", "720", "480", "360"}
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern+"p") || strings.Contains(lower, pattern+"x") || strings.Contains(lower, "_"+pattern) || strings.Contains(lower, "-"+pattern) {
			return pattern + "p"
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	for key, values := range parsed.Query() {
		keyLower := strings.ToLower(key)
		if keyLower != "quality" && keyLower != "q" && keyLower != "resolution" {
			continue
		}
		for _, value := range values {
			if normalized := normalizeQuality(value); normalized != "" {
				return normalized
			}
		}
	}
	return ""
}

// probeM3U8Quality 会尽量从 master playlist 中提取分辨率信息。
// 这一步只是给线路页一个可读提示，不应该因为探测失败而影响播放。
func probeM3U8Quality(client *http.Client, playlistURL string) (string, error) {
	resp, err := client.Get(playlistURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}

	content := string(body)
	if !strings.Contains(content, "#EXTM3U") {
		return "", nil
	}

	re := regexp.MustCompile(`RESOLUTION=\d+x(\d+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	bestHeight := 0
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		value, err := strconv.Atoi(match[1])
		if err == nil && value > bestHeight {
			bestHeight = value
		}
	}
	if bestHeight > 0 {
		return fmt.Sprintf("自适应(最高%dp)", bestHeight), nil
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if normalized := normalizeQuality(line); normalized != "" {
			return normalized, nil
		}
	}
	return "", nil
}

func normalizeQuality(value string) string {
	re := regexp.MustCompile(`(?i)(2160|1440|1080|720|480|360)\s*p?`)
	match := re.FindStringSubmatch(value)
	if len(match) >= 2 {
		return match[1] + "p"
	}
	return ""
}
