package source

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// GroupEpisodes 按集数归类剧集，合并重复线路
func GroupEpisodes(episodes []Episode) []EpisodeGroup {
	if len(episodes) == 0 {
		return nil
	}

	groups := make([]EpisodeGroup, 0, len(episodes))
	indexByKey := make(map[string]int, len(episodes))
	seenURLs := make(map[string]map[string]struct{}, len(episodes))

	for i, episode := range episodes {
		key := episodeGroupKey(episode)
		if idx, ok := indexByKey[key]; ok {
			if _, exists := seenURLs[key][episode.URL]; exists {
				continue
			}
			seenURLs[key][episode.URL] = struct{}{}
			groups[idx].Candidates = append(groups[idx].Candidates, EpisodeCandidate{
				Name: episode.Name,
				URL:  episode.URL,
			})
			groups[idx].Name = betterEpisodeName(groups[idx].Name, episode.Name)
			continue
		}

		indexByKey[key] = len(groups)
		seenURLs[key] = map[string]struct{}{
			episode.URL: {},
		}
		groups = append(groups, EpisodeGroup{
			Name:        episode.Name,
			Number:      episode.Number,
			NumberValue: episode.NumberValue,
			HasNumber:   episode.HasNumber,
			Candidates: []EpisodeCandidate{
				{
					Name: episode.Name,
					URL:  episode.URL,
				},
			},
			order: i,
		})
	}

	sort.SliceStable(groups, func(i, j int) bool {
		left := groups[i]
		right := groups[j]

		if left.HasNumber && right.HasNumber {
			if left.NumberValue != right.NumberValue {
				return left.NumberValue < right.NumberValue
			}
			return left.order < right.order
		}
		if left.HasNumber != right.HasNumber {
			return left.HasNumber
		}
		return left.order < right.order
	})

	return groups
}

func episodeGroupKey(episode Episode) string {
	if episode.HasNumber && episode.Number != "" {
		return "n:" + normalizeNumberString(episode.Number)
	}
	return "t:" + normalizeEpisodeText(episode.Name)
}

func normalizeEpisodeText(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "\t", "")
	name = strings.ReplaceAll(name, "\n", "")
	return name
}

func normalizeNumberString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return strconv.FormatFloat(parsed, 'f', -1, 64)
	}
	return value
}

func betterEpisodeName(current, candidate string) string {
	current = strings.TrimSpace(current)
	candidate = strings.TrimSpace(candidate)
	if current == "" {
		return candidate
	}
	if candidate == "" {
		return current
	}

	currentScore := episodeNameScore(current)
	candidateScore := episodeNameScore(candidate)
	if candidateScore > currentScore {
		return candidate
	}
	if candidateScore == currentScore && len(candidate) > len(current) {
		return candidate
	}
	return current
}

func episodeNameScore(name string) int {
	score := 0
	if strings.Contains(name, "第") {
		score += 2
	}
	if strings.Contains(name, "集") || strings.Contains(name, "话") {
		score += 2
	}
	if _, err := strconv.ParseFloat(strings.TrimSpace(name), 64); err != nil {
		score++
	}
	return score
}

// Label 返回适合展示的剧集名称
func (g EpisodeGroup) Label() string {
	if len(g.Candidates) > 1 {
		return fmt.Sprintf("%s  (%d 条线路)", g.Name, len(g.Candidates))
	}
	return g.Name
}
