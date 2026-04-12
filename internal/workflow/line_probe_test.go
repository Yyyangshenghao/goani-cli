package workflow

import "testing"

func TestBuildLineSelectionItemsMovesFailedCandidatesToEnd(t *testing.T) {
	items := buildLineSelectionItems([]resolvedEpisodeCandidate{
		{
			name:       "失败线路",
			episodeURL: "https://example.com/ep-fail",
			err:        errString("解析失败"),
		},
		{
			name:       "可播线路A",
			episodeURL: "https://example.com/ep-a",
			videoURL:   "https://example.com/a.m3u8",
			format:     "m3u8",
			quality:    "1080p",
		},
		{
			name:       "可播线路B",
			episodeURL: "https://example.com/ep-b",
			videoURL:   "https://example.com/b.mp4",
			format:     "mp4",
			quality:    "720p",
		},
	})

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	if items[0].Title != "可播线路A" || items[1].Title != "可播线路B" {
		t.Fatalf("expected playable items to stay at the front in original order, got %q then %q", items[0].Title, items[1].Title)
	}
	if items[2].Title != "失败线路" {
		t.Fatalf("expected failed item to move to the end, got %q", items[2].Title)
	}
	if items[2].Error == "" {
		t.Fatalf("expected failed item to keep its error message")
	}
}

type errString string

func (e errString) Error() string {
	return string(e)
}
