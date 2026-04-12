package workflow

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
	consoleui "github.com/Yyyangshenghao/goani-cli/internal/ui/console"
)

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
	candidates := sortedCandidatesByPriority(application, group.Candidates)
	for i, candidate := range candidates {
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

func sortedCandidatesByPriority(application *app.App, candidates []source.EpisodeCandidate) []source.EpisodeCandidate {
	sorted := make([]source.EpisodeCandidate, len(candidates))
	copy(sorted, candidates)
	sort.SliceStable(sorted, func(i, j int) bool {
		leftPriority := application.SourceManager.GetChannelPriorityByName(sorted[i].SourceName)
		rightPriority := application.SourceManager.GetChannelPriorityByName(sorted[j].SourceName)
		if leftPriority != rightPriority {
			return leftPriority > rightPriority
		}
		return sorted[i].SourceName < sorted[j].SourceName
	})
	return sorted
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
