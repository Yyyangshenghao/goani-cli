package webselector

import "testing"

func TestParseEpisodeNumberRejectsEpisodeRanges(t *testing.T) {
	number, value, ok := parseEpisodeNumber("第01-10集", "")
	if ok {
		t.Fatalf("expected range title to be treated as non-single episode, got ok with number=%q value=%v", number, value)
	}
	if number != "" || value != 0 {
		t.Fatalf("expected empty parsed number for range title, got number=%q value=%v", number, value)
	}
}

func TestParseEpisodeNumberKeepsSingleEpisodeTitles(t *testing.T) {
	number, value, ok := parseEpisodeNumber("第01集", "")
	if !ok {
		t.Fatalf("expected single episode title to parse successfully")
	}
	if number != "1" || value != 1 {
		t.Fatalf("unexpected parsed result: number=%q value=%v", number, value)
	}
}
