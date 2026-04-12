package app

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

const (
	sourceDoctorTimeout            = 6 * time.Second
	sourceDoctorConcurrency        = 4
	sourceDoctorCaseConcurrency    = 3
	sourceDoctorMaxAnimeCandidates = 2
	sourceDoctorMaxEpisodeChecks   = 2
)

var defaultDoctorKeywords = []string{
	"葬送的芙莉莲",
	"进击的巨人",
	"鬼灭之刃",
	"海贼王",
	"名侦探柯南",
}

var doctorCaseRunner = runDoctorCase

// SourceDoctorResult 表示单个渠道的一次 doctor 诊断结果。
type SourceDoctorResult struct {
	ID       string
	Name     string
	Enabled  bool
	Priority int
	Snapshot source.SourceDoctorSnapshot
}

// RunSourceDoctor 对当前所有渠道执行固定样本诊断，并自动关闭 0 成功渠道、提高高成功渠道优先级。
func (a *App) RunSourceDoctor() ([]SourceDoctorResult, error) {
	loadedSources := a.SourceManager.GetAll()
	if len(loadedSources) == 0 {
		return nil, fmt.Errorf("未找到媒体源")
	}

	results := make([]SourceDoctorResult, len(loadedSources))
	updates := make([]source.ChannelDoctorUpdate, len(loadedSources))
	limiter := make(chan struct{}, sourceDoctorConcurrency)
	var wg sync.WaitGroup

	for i, ms := range loadedSources {
		wg.Add(1)
		go func(index int, mediaSource source.MediaSource) {
			defer wg.Done()
			limiter <- struct{}{}
			defer func() { <-limiter }()

			result := runDoctorForSource(mediaSource)
			results[index] = result
			updates[index] = source.ChannelDoctorUpdate{
				ID:       result.ID,
				Name:     result.Name,
				Priority: result.Priority,
				Enabled:  result.Enabled,
				Snapshot: result.Snapshot,
			}
		}(i, ms)
	}

	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Priority != results[j].Priority {
			return results[i].Priority > results[j].Priority
		}
		return results[i].Name < results[j].Name
	})

	if err := a.SourceManager.UpdateDoctorSnapshots(updates); err != nil {
		return results, err
	}
	return results, nil
}

func runDoctorForSource(mediaSource source.MediaSource) SourceDoctorResult {
	startedAt := time.Now()
	result := SourceDoctorResult{
		ID:   mediaSource.PreferenceID(),
		Name: doctorChannelName(mediaSource),
	}

	ws := webselector.New(mediaSource).WithTimeout(sourceDoctorTimeout)
	cases := make([]source.SourceDoctorCase, len(defaultDoctorKeywords))
	caseLimiter := make(chan struct{}, sourceDoctorCaseConcurrency)
	var wg sync.WaitGroup

	for i, keyword := range defaultDoctorKeywords {
		wg.Add(1)
		go func(index int, keyword string) {
			defer wg.Done()
			caseLimiter <- struct{}{}
			defer func() { <-caseLimiter }()

			cases[index] = doctorCaseRunner(ws, keyword)
		}(i, keyword)
	}
	wg.Wait()

	successCount := 0
	var totalDuration int64
	for _, item := range cases {
		totalDuration += item.DurationMS
		if item.Status == source.DoctorStatusVideoReady {
			successCount++
		}
	}

	averageMS := int64(0)
	if len(cases) > 0 {
		averageMS = totalDuration / int64(len(cases))
	}

	priority := buildDoctorPriority(successCount, averageMS)
	enabled := successCount > 0
	result.Enabled = enabled
	result.Priority = priority
	result.Snapshot = source.SourceDoctorSnapshot{
		CheckedAt:      startedAt.Format("2006-01-02 15:04:05"),
		SuccessfulRuns: successCount,
		TotalRuns:      len(defaultDoctorKeywords),
		AverageMS:      averageMS,
		Priority:       priority,
		AutoDisabled:   !enabled,
		Summary:        doctorSummary(successCount, len(defaultDoctorKeywords), averageMS, enabled),
		Cases:          cases,
	}

	return result
}

func runDoctorCase(ws *webselector.WebSelectorSource, keyword string) source.SourceDoctorCase {
	startedAt := time.Now()
	result := source.SourceDoctorCase{
		CheckedAt: startedAt.Format("2006-01-02 15:04:05"),
		Keyword:   keyword,
		Status:    source.DoctorStatusUnknown,
	}

	searchResults, err := ws.Search(keyword)
	if err != nil {
		result.Status = source.DoctorStatusSearchError
		result.Message = fmt.Sprintf("搜索失败: %v", err)
		result.DurationMS = time.Since(startedAt).Milliseconds()
		return result
	}

	result.SearchHits = len(searchResults)
	if len(searchResults) == 0 {
		result.Status = source.DoctorStatusNoSearchHits
		result.Message = "搜索无结果"
		result.DurationMS = time.Since(startedAt).Milliseconds()
		return result
	}

	searchLimit := minInt(len(searchResults), sourceDoctorMaxAnimeCandidates)
	var lastEpisodeErr error
	var lastVideoErr error

	for _, anime := range searchResults[:searchLimit] {
		episodes, err := ws.GetEpisodes(anime.URL)
		if err != nil {
			lastEpisodeErr = err
			continue
		}
		if len(episodes) == 0 {
			continue
		}

		result.EpisodeCount = len(episodes)
		episodeChecks := minInt(len(episodes), sourceDoctorMaxEpisodeChecks)
		for _, episode := range episodes[:episodeChecks] {
			videoURL, err := ws.GetVideoURL(episode.URL)
			if err != nil {
				lastVideoErr = err
				continue
			}
			if strings.TrimSpace(videoURL) == "" {
				continue
			}

			result.Status = source.DoctorStatusVideoReady
			result.VideoReady = true
			result.Message = "拿到视频直链"
			result.DurationMS = time.Since(startedAt).Milliseconds()
			return result
		}
	}

	switch {
	case result.EpisodeCount == 0 && lastEpisodeErr != nil:
		result.Status = source.DoctorStatusEpisodeError
		result.Message = fmt.Sprintf("剧集解析失败: %v", lastEpisodeErr)
	case result.EpisodeCount == 0:
		result.Status = source.DoctorStatusNoEpisodes
		result.Message = "搜到条目，但没有拿到剧集"
	case lastVideoErr != nil:
		result.Status = source.DoctorStatusVideoError
		result.Message = fmt.Sprintf("直链解析失败: %v", lastVideoErr)
	default:
		result.Status = source.DoctorStatusVideoError
		result.Message = "拿到剧集，但没有解析出视频直链"
	}
	result.DurationMS = time.Since(startedAt).Milliseconds()
	return result
}

func doctorSummary(successCount, totalCount int, averageMS int64, enabled bool) string {
	status := "已保留"
	if !enabled {
		status = "已自动关闭"
	}
	return fmt.Sprintf("5 项样本成功 %d/%d，平均 %dms，%s", successCount, totalCount, averageMS, status)
}

func buildDoctorPriority(successCount int, averageMS int64) int {
	if successCount <= 0 {
		return 0
	}
	latencyBonus := 10000 - int(averageMS)
	if latencyBonus < 0 {
		latencyBonus = 0
	}
	return successCount*100000 + latencyBonus
}

func doctorChannelName(mediaSource source.MediaSource) string {
	name := strings.TrimSpace(mediaSource.Arguments.Name)
	if name != "" {
		return name
	}
	return strings.TrimSpace(mediaSource.FactoryID)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
