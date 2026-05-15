package mcp

import (
	"encoding/json"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
}

func TestSearchInputDefault(t *testing.T) {
	in := searchInput{}
	if in.Keyword != "" {
		t.Errorf("searchInput.Keyword default = %q, want empty string", in.Keyword)
	}
}

func TestGetEpisodesInputDefault(t *testing.T) {
	in := getEpisodesInput{}
	if in.URL != "" {
		t.Errorf("getEpisodesInput.URL default = %q, want empty string", in.URL)
	}
}

func TestGetVideoURLInputDefault(t *testing.T) {
	in := getVideoURLInput{}
	if in.EpisodeURL != "" {
		t.Errorf("getVideoURLInput.EpisodeURL default = %q, want empty string", in.EpisodeURL)
	}
}

func TestMarshalJSONSearchInput(t *testing.T) {
	in := searchInput{Keyword: "test"}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("json.Marshal(searchInput) error: %v", err)
	}
	want := `{"keyword":"test"}`
	if string(data) != want {
		t.Errorf("json.Marshal(searchInput) = %s, want %s", data, want)
	}
}

func TestMarshalJSONAnimeItem(t *testing.T) {
	item := animeItem{Name: "Naruto", URL: "https://example.com/naruto"}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal(animeItem) error: %v", err)
	}
	want := `{"name":"Naruto","url":"https://example.com/naruto"}`
	if string(data) != want {
		t.Errorf("json.Marshal(animeItem) = %s, want %s", data, want)
	}
}

func TestMarshalJSONEpisodeItem(t *testing.T) {
	item := episodeItem{Name: "EP01", URL: "https://example.com/ep1"}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal(episodeItem) error: %v", err)
	}
	want := `{"name":"EP01","url":"https://example.com/ep1"}`
	if string(data) != want {
		t.Errorf("json.Marshal(episodeItem) = %s, want %s", data, want)
	}
}

func TestMarshalJSONSourceItem(t *testing.T) {
	item := sourceItem{ID: "src-1", Name: "test-source", Description: "desc", Enabled: true}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal(sourceItem) error: %v", err)
	}
	want := `{"id":"src-1","name":"test-source","description":"desc","enabled":true}`
	if string(data) != want {
		t.Errorf("json.Marshal(sourceItem) = %s, want %s", data, want)
	}
}
