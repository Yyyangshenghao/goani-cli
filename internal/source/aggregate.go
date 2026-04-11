package source

import (
	"sort"
	"strings"
	"unicode"
)

// AnimeHit 表示某个媒体源返回的一条番剧搜索命中。
type AnimeHit struct {
	SourceName string
	Anime      Anime
}

// AggregatedAnime 表示按标题聚合后的番剧结果。
type AggregatedAnime struct {
	Name string
	Hits []AnimeHit

	key   string
	order int
}

// GroupAnimes 按归一化标题聚合多个媒体源的搜索结果。
func GroupAnimes(hits []AnimeHit) []AggregatedAnime {
	if len(hits) == 0 {
		return nil
	}

	groups := make([]AggregatedAnime, 0, len(hits))
	indexByKey := make(map[string]int, len(hits))
	seenHits := make(map[string]map[string]struct{}, len(hits))

	for i, hit := range hits {
		key := animeGroupKey(hit.Anime.Name)
		dedupeKey := hit.SourceName + "\n" + strings.TrimSpace(hit.Anime.URL)

		if idx, ok := indexByKey[key]; ok {
			if _, exists := seenHits[key][dedupeKey]; exists {
				continue
			}
			seenHits[key][dedupeKey] = struct{}{}
			groups[idx].Hits = append(groups[idx].Hits, hit)
			groups[idx].Name = betterAnimeName(groups[idx].Name, hit.Anime.Name)
			continue
		}

		indexByKey[key] = len(groups)
		seenHits[key] = map[string]struct{}{
			dedupeKey: {},
		}
		groups = append(groups, AggregatedAnime{
			Name:  strings.TrimSpace(hit.Anime.Name),
			Hits:  []AnimeHit{hit},
			key:   key,
			order: i,
		})
	}

	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].order < groups[j].order
	})

	return groups
}

// SourceCount 返回当前聚合结果覆盖的媒体源数量。
func (a AggregatedAnime) SourceCount() int {
	if len(a.Hits) == 0 {
		return 0
	}

	seen := make(map[string]struct{}, len(a.Hits))
	for _, hit := range a.Hits {
		if strings.TrimSpace(hit.SourceName) == "" {
			continue
		}
		seen[hit.SourceName] = struct{}{}
	}
	if len(seen) == 0 {
		return len(a.Hits)
	}
	return len(seen)
}

// HitCount 返回合并到当前番剧下的命中条数。
func (a AggregatedAnime) HitCount() int {
	return len(a.Hits)
}

func animeGroupKey(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return name
	}

	var b strings.Builder
	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		b.WriteRune(r)
	}

	normalized := b.String()
	if normalized == "" {
		return name
	}
	return normalized
}

func betterAnimeName(current, candidate string) string {
	current = strings.TrimSpace(current)
	candidate = strings.TrimSpace(candidate)
	if current == "" {
		return candidate
	}
	if candidate == "" {
		return current
	}
	if len([]rune(candidate)) < len([]rune(current)) {
		return candidate
	}
	return current
}
