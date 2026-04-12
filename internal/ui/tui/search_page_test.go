package tui

import "testing"

func TestSearchResultPageSizeCapsAtTen(t *testing.T) {
	model := searchTUIModel{height: 40}

	if got := model.resultPageSize(); got != 10 {
		t.Fatalf("unexpected page size: got %d want %d", got, 10)
	}
}

func TestSearchResultPageSizeFallsBackForUnknownHeight(t *testing.T) {
	model := searchTUIModel{}

	if got := model.resultPageSize(); got != 5 {
		t.Fatalf("unexpected fallback page size: got %d want %d", got, 5)
	}
}

func TestSearchVisibleResultRangeCentersAroundSelection(t *testing.T) {
	model := searchTUIModel{
		height:   18,
		selected: 6,
	}

	start, end := model.visibleResultRange(20)
	if start != 2 || end != 11 {
		t.Fatalf("unexpected visible range: got [%d,%d) want [%d,%d)", start, end, 2, 11)
	}
}

func TestSearchVisibleResultRangeClampsAtEnd(t *testing.T) {
	model := searchTUIModel{
		height:   20,
		selected: 14,
	}

	start, end := model.visibleResultRange(15)
	if start != 5 || end != 15 {
		t.Fatalf("unexpected clamped range: got [%d,%d) want [%d,%d)", start, end, 5, 15)
	}
}
