package player

import "testing"

func TestMPVCommandArgsPutOptionsBeforeURL(t *testing.T) {
	args := mpvCommandArgs("https://media.example.com/master.m3u8", []string{
		"--user-agent=test-agent",
		"--referrer=https://anime.example.com/watch/1",
	})

	expected := []string{
		"--user-agent=test-agent",
		"--referrer=https://anime.example.com/watch/1",
		"https://media.example.com/master.m3u8",
	}
	if len(args) != len(expected) {
		t.Fatalf("unexpected arg length: got %d want %d", len(args), len(expected))
	}
	for i := range expected {
		if args[i] != expected[i] {
			t.Fatalf("unexpected arg at %d: got %q want %q", i, args[i], expected[i])
		}
	}
}
