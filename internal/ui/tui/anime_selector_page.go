package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
)

// RunAnimeSelectionTUI 运行番剧选择 TUI
func RunAnimeSelectionTUI(sourceName string, animes []source.Anime) (*source.Anime, error) {
	items := make([]string, len(animes))
	jumpValues := make([]string, len(animes))
	for i, anime := range animes {
		items[i] = anime.Name
		jumpValues[i] = strconv.Itoa(i + 1)
	}

	filterInput := textinput.New()
	filterInput.Placeholder = "在当前结果里继续搜索"
	filterInput.CharLimit = 64
	filterInput.Width = 36
	filterInput.Focus()

	model := selectorTUIModel{
		title:        fmt.Sprintf("片源: %s", sourceName),
		subtitle:     "输入关键字过滤当前结果，Enter 确认，Esc 返回片源列表",
		emptyMessage: "当前片源没有可选番剧",
		items:        items,
		jumpValues:   jumpValues,
		showOrdinal:  true,
		allowFilter:  true,
		filterInput:  filterInput,
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, err
	}

	return &animes[result.index], nil
}

// RunAggregatedAnimeSelectionTUI 运行跨片源聚合后的番剧选择页。
func RunAggregatedAnimeSelectionTUI(animes []source.AggregatedAnime, initialIndex int) (*source.AggregatedAnime, int, error) {
	items := make([]string, len(animes))
	jumpValues := make([]string, len(animes))
	for i, anime := range animes {
		items[i] = aggregatedAnimeSelectionLabel(anime)
		jumpValues[i] = strconv.Itoa(i + 1)
	}

	filterInput := textinput.New()
	filterInput.Placeholder = "在聚合结果里继续搜索"
	filterInput.CharLimit = 64
	filterInput.Width = 36
	filterInput.Focus()

	model := selectorTUIModel{
		title:        "聚合番剧",
		subtitle:     "输入关键字过滤当前结果，Enter 确认，Esc 返回上一层",
		emptyMessage: "当前搜索结果没有可选番剧",
		items:        items,
		jumpValues:   jumpValues,
		showOrdinal:  true,
		allowFilter:  true,
		filterInput:  filterInput,
		selected:     clampSelectionIndex(initialIndex, len(animes)),
	}

	result, err := runSelectorTUI(model)
	if err != nil || result == nil {
		return nil, -1, err
	}

	return &animes[result.index], result.index, nil
}

func aggregatedAnimeSelectionLabel(anime source.AggregatedAnime) string {
	return fmt.Sprintf("%s  (%d 个片源 / %d 条命中)", anime.Name, anime.SourceCount(), anime.HitCount())
}
