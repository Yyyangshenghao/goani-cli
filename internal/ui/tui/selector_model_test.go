package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectorReverseMovesSelectionToTopOfReversedOrder(t *testing.T) {
	model := selectorTUIModel{
		items:        []string{"1", "2", "3"},
		jumpValues:   []string{"1", "2", "3"},
		allowReverse: true,
		selected:     1,
	}
	model.applyFilter()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	got, ok := updated.(selectorTUIModel)
	if !ok {
		t.Fatalf("expected selectorTUIModel, got %T", updated)
	}

	if !got.reversed {
		t.Fatalf("expected reversed mode to be enabled")
	}
	if got.selected != 0 {
		t.Fatalf("expected selected display index to reset to 0, got %d", got.selected)
	}
	if actual := got.displayIndexToActual(got.selected); actual != 2 {
		t.Fatalf("expected reversed top item to point to last episode, got actual index %d", actual)
	}
}

func TestSelectorReverseHighlightsLastEpisodeInView(t *testing.T) {
	model := selectorTUIModel{
		title:        "剧集",
		subtitle:     "test",
		items:        []string{"1", "2", "3"},
		jumpValues:   []string{"1", "2", "3"},
		allowReverse: true,
		selected:     0,
		windowHeight: 20,
	}
	model.applyFilter()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	got := updated.(selectorTUIModel)
	view := got.View()

	if !strings.Contains(view, "> 3") {
		t.Fatalf("expected view to highlight last episode after reverse, got:\n%s", view)
	}
}
