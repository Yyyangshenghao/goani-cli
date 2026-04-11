package tui

import "testing"

func TestBuildPlaybackMenuItemsIncludesAdjacentEpisodeActions(t *testing.T) {
	items := buildPlaybackMenuItems(true, true)

	if len(items) < 2 {
		t.Fatalf("expected previous and next actions to be present")
	}
	if items[0].action != PlaybackActionPreviousEpisode {
		t.Fatalf("expected first action to be previous episode, got %q", items[0].action)
	}
	if items[1].action != PlaybackActionNextEpisode {
		t.Fatalf("expected second action to be next episode, got %q", items[1].action)
	}
}

func TestBuildPlaybackMenuItemsSkipsUnavailableAdjacentEpisodeActions(t *testing.T) {
	items := buildPlaybackMenuItems(false, true)

	if len(items) == 0 {
		t.Fatalf("expected menu items to be present")
	}
	if items[0].action != PlaybackActionNextEpisode {
		t.Fatalf("expected next episode to be the first action when previous is unavailable, got %q", items[0].action)
	}

	for _, item := range items {
		if item.action == PlaybackActionPreviousEpisode {
			t.Fatalf("did not expect previous episode action when previous is unavailable")
		}
	}
}
