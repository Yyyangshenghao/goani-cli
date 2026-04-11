package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PlaybackAction 表示播放状态页接下来要回到哪一层。
type PlaybackAction string

const (
	PlaybackActionPreviousEpisode PlaybackAction = "previous_episode"
	PlaybackActionNextEpisode     PlaybackAction = "next_episode"
	PlaybackActionEpisodeList     PlaybackAction = "episode_list"
	PlaybackActionAnimeList       PlaybackAction = "anime_list"
	PlaybackActionHome            PlaybackAction = "home"
	PlaybackActionQuit            PlaybackAction = "quit"
)

type playbackMenuItem struct {
	title       string
	description string
	action      PlaybackAction
}

// playbackPageModel 管理播放器启动后的状态页。
type playbackPageModel struct {
	animeName    string
	episodeLabel string
	playerName   string
	lineLabel    string
	items        []playbackMenuItem
	selected     int
	action       PlaybackAction
}

// RunPlaybackPageTUI 展示“播放器已启动”页面，避免成功播放后直接退出整个 TUI。
func RunPlaybackPageTUI(animeName, episodeLabel, playerName, lineLabel string, hasPrevious, hasNext bool) (PlaybackAction, error) {
	model := playbackPageModel{
		animeName:    animeName,
		episodeLabel: episodeLabel,
		playerName:   playerName,
		lineLabel:    lineLabel,
		items:        buildPlaybackMenuItems(hasPrevious, hasNext),
	}

	finalModel, err := newProgram(model).Run()
	if err != nil {
		return PlaybackActionQuit, err
	}

	result, ok := finalModel.(playbackPageModel)
	if !ok {
		return PlaybackActionQuit, fmt.Errorf("无法读取播放状态页")
	}
	if result.action == "" {
		return PlaybackActionEpisodeList, nil
	}
	return result.action, nil
}

func (m playbackPageModel) Init() tea.Cmd {
	return nil
}

func (m playbackPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.action = PlaybackActionQuit
			return m, tea.Quit
		case "esc":
			m.action = PlaybackActionEpisodeList
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "enter":
			m.action = m.items[m.selected].action
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m playbackPageModel) View() string {
	var b strings.Builder

	b.WriteString(tuiTitleStyle.Render("正在播放"))
	b.WriteString("\n")
	b.WriteString(tuiOkStyle.Render(fmt.Sprintf("播放器已启动: %s", m.playerName)))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("番剧: %s\n", m.animeName))
	b.WriteString(fmt.Sprintf("剧集: %s\n", m.episodeLabel))
	if strings.TrimSpace(m.lineLabel) != "" {
		b.WriteString(fmt.Sprintf("线路: %s\n", m.lineLabel))
	}
	b.WriteString("\n")

	for i, item := range m.items {
		line := fmt.Sprintf("%s %s", pointer(i == m.selected), item.title)
		if i == m.selected {
			b.WriteString(tuiPickStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(tuiMutedStyle.Render(item.description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(tuiHintStyle.Render("↑/↓ 选择，Enter 确认，Esc 返回选集，q 退出"))

	return strings.TrimRight(b.String(), "\n")
}

func buildPlaybackMenuItems(hasPrevious, hasNext bool) []playbackMenuItem {
	items := make([]playbackMenuItem, 0, 6)
	if hasPrevious {
		items = append(items, playbackMenuItem{
			title:       "上一集",
			description: "直接跳到当前番剧的上一集",
			action:      PlaybackActionPreviousEpisode,
		})
	}
	if hasNext {
		items = append(items, playbackMenuItem{
			title:       "下一集",
			description: "直接跳到当前番剧的下一集",
			action:      PlaybackActionNextEpisode,
		})
	}

	items = append(items,
		playbackMenuItem{
			title:       "返回选集",
			description: "继续留在当前番剧的选集页",
			action:      PlaybackActionEpisodeList,
		},
		playbackMenuItem{
			title:       "返回番剧列表",
			description: "回到当前片源的番剧列表",
			action:      PlaybackActionAnimeList,
		},
		playbackMenuItem{
			title:       "返回首页",
			description: "回到 goani TUI 首页",
			action:      PlaybackActionHome,
		},
		playbackMenuItem{
			title:       "退出",
			description: "退出 goani，保留外部播放器继续播放",
			action:      PlaybackActionQuit,
		},
	)

	return items
}
