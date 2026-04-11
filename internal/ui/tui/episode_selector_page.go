package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
)

// RunEpisodeSelectionTUI 运行剧集选择 TUI
func RunEpisodeSelectionTUI(animeName string, episodes []source.EpisodeGroup) (*source.EpisodeGroup, error) {
	result, _, err := RunEpisodeSelectionTUIWithSelection(animeName, episodes, 0)
	return result, err
}

// RunEpisodeSelectionTUIWithSelection 运行剧集选择 TUI，并允许调用方指定返回后默认高亮的剧集。
func RunEpisodeSelectionTUIWithSelection(animeName string, episodes []source.EpisodeGroup, initialIndex int) (*source.EpisodeGroup, int, error) {
	items := make([]string, len(episodes))
	jumpValues := make([]string, len(episodes))
	for i, episode := range episodes {
		items[i] = episodeSelectionLabel(episode)
		if episode.HasNumber && episode.Number != "" {
			jumpValues[i] = normalizeDigitInput(episode.Number)
		} else {
			jumpValues[i] = strconv.Itoa(i + 1)
		}
	}

	model := selectorTUIModel{
		title:        fmt.Sprintf("剧集: %s", animeName),
		subtitle:     "Enter 播放，r 切换顺序/倒序，输入集数跳转，Esc 返回番剧列表",
		emptyMessage: "当前番剧没有可选剧集",
		items:        items,
		jumpValues:   jumpValues,
		showOrdinal:  false,
		allowReverse: true,
		allowNumber:  true,
		reversed:     false,
		selected:     clampSelectionIndex(initialIndex, len(episodes)),
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, -1, err
	}

	return &episodes[result.index], result.index, nil
}

// episodeSelectionLabel 让选集页优先展示真实集号，而不是再额外叠加一层列表编号。
func episodeSelectionLabel(episode source.EpisodeGroup) string {
	label := strings.TrimSpace(episode.Name)
	if episode.HasNumber && strings.TrimSpace(episode.Number) != "" {
		label = normalizeDigitInput(episode.Number)
	}
	if len(episode.Candidates) > 1 {
		return fmt.Sprintf("%s  (%d 条线路)", label, len(episode.Candidates))
	}
	return label
}
