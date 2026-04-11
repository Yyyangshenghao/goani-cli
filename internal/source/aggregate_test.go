package source

import "testing"

func TestGroupAnimesMergesHitsAcrossSources(t *testing.T) {
	hits := []AnimeHit{
		{
			SourceName: "源A",
			Anime: Anime{
				Name: "葬送的芙莉莲",
				URL:  "https://a.example.com/frieren",
			},
		},
		{
			SourceName: "源B",
			Anime: Anime{
				Name: "葬送 的 芙莉莲",
				URL:  "https://b.example.com/frieren",
			},
		},
		{
			SourceName: "源C",
			Anime: Anime{
				Name: "迷宫饭",
				URL:  "https://c.example.com/dungeon",
			},
		},
	}

	groups := GroupAnimes(hits)

	if len(groups) != 2 {
		t.Fatalf("unexpected group count: got %d want %d", len(groups), 2)
	}
	if groups[0].Name != "葬送的芙莉莲" {
		t.Fatalf("unexpected aggregated anime name: got %q", groups[0].Name)
	}
	if groups[0].SourceCount() != 2 {
		t.Fatalf("unexpected source count: got %d want %d", groups[0].SourceCount(), 2)
	}
	if groups[0].HitCount() != 2 {
		t.Fatalf("unexpected hit count: got %d want %d", groups[0].HitCount(), 2)
	}
}

func TestGroupAnimesDeduplicatesSameSourceAndURL(t *testing.T) {
	hits := []AnimeHit{
		{
			SourceName: "源A",
			Anime: Anime{
				Name: "孤独摇滚",
				URL:  "https://a.example.com/bocchi",
			},
		},
		{
			SourceName: "源A",
			Anime: Anime{
				Name: "孤独摇滚!",
				URL:  "https://a.example.com/bocchi",
			},
		},
	}

	groups := GroupAnimes(hits)

	if len(groups) != 1 {
		t.Fatalf("unexpected group count: got %d want %d", len(groups), 1)
	}
	if groups[0].HitCount() != 1 {
		t.Fatalf("unexpected hit count after dedupe: got %d want %d", groups[0].HitCount(), 1)
	}
}
