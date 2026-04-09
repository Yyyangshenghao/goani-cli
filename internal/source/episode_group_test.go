package source

import "testing"

func TestGroupEpisodesMergesSameEpisodeNumberAcrossDifferentNames(t *testing.T) {
	episodes := []Episode{
		{Name: "第1集", URL: "https://a.example.com/1", Number: "1", NumberValue: 1, HasNumber: true},
		{Name: "1", URL: "https://b.example.com/1", Number: "01", NumberValue: 1, HasNumber: true},
		{Name: "01", URL: "https://c.example.com/1", Number: "1.0", NumberValue: 1, HasNumber: true},
		{Name: "第2集", URL: "https://a.example.com/2", Number: "2", NumberValue: 2, HasNumber: true},
	}

	groups := GroupEpisodes(episodes)

	if len(groups) != 2 {
		t.Fatalf("unexpected group count: got %d want %d", len(groups), 2)
	}
	if groups[0].Name != "第1集" {
		t.Fatalf("unexpected preferred group name: got %q want %q", groups[0].Name, "第1集")
	}
	if len(groups[0].Candidates) != 3 {
		t.Fatalf("unexpected candidate count for episode 1: got %d want %d", len(groups[0].Candidates), 3)
	}
	if groups[0].Label() != "第1集  (3 条线路)" {
		t.Fatalf("unexpected label: got %q", groups[0].Label())
	}
	if groups[1].Name != "第2集" {
		t.Fatalf("unexpected second group name: got %q", groups[1].Name)
	}
}

func TestGroupEpisodesDeduplicatesSameCandidateURL(t *testing.T) {
	episodes := []Episode{
		{Name: "第13.5集", URL: "https://a.example.com/13.5", Number: "13.5", NumberValue: 13.5, HasNumber: true},
		{Name: "13.5", URL: "https://a.example.com/13.5", Number: "13.5", NumberValue: 13.5, HasNumber: true},
		{Name: "13.5", URL: "https://b.example.com/13.5", Number: "13.5", NumberValue: 13.5, HasNumber: true},
	}

	groups := GroupEpisodes(episodes)

	if len(groups) != 1 {
		t.Fatalf("unexpected group count: got %d want %d", len(groups), 1)
	}
	if len(groups[0].Candidates) != 2 {
		t.Fatalf("unexpected deduplicated candidate count: got %d want %d", len(groups[0].Candidates), 2)
	}
	if groups[0].Number != "13.5" {
		t.Fatalf("unexpected episode number: got %q want %q", groups[0].Number, "13.5")
	}
}
