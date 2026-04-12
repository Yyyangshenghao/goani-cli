package player

import (
	"net/http"
	"testing"
)

func TestEncodeDecodeStreamRequestContextRoundTrip(t *testing.T) {
	original := StreamRequestContext{
		SourceURL: " https://media.example.com/master.m3u8 ",
		Referer:   " https://anime.example.com/watch/1 ",
		UserAgent: " goani-test-agent ",
		Cookies:   " session=abc123 ",
		Headers: map[string]string{
			"X-Trace-ID": " trace-001 ",
			" ":          "ignored",
		},
	}

	encoded, err := EncodeStreamRequestContext(original)
	if err != nil {
		t.Fatalf("EncodeStreamRequestContext returned error: %v", err)
	}

	decoded, err := DecodeStreamRequestContext(encoded)
	if err != nil {
		t.Fatalf("DecodeStreamRequestContext returned error: %v", err)
	}

	if decoded.SourceURL != "https://media.example.com/master.m3u8" {
		t.Fatalf("unexpected source url: got %q", decoded.SourceURL)
	}
	if decoded.Referer != "https://anime.example.com/watch/1" {
		t.Fatalf("unexpected referer: got %q", decoded.Referer)
	}
	if decoded.UserAgent != "goani-test-agent" {
		t.Fatalf("unexpected user agent: got %q", decoded.UserAgent)
	}
	if decoded.Cookies != "session=abc123" {
		t.Fatalf("unexpected cookies: got %q", decoded.Cookies)
	}
	if decoded.Headers["X-Trace-ID"] != "trace-001" {
		t.Fatalf("unexpected custom header: got %q", decoded.Headers["X-Trace-ID"])
	}
}

func TestApplyStreamRequestContextAllowsHeaderOverrides(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://media.example.com/master.m3u8", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	applyStreamRequestContext(req, StreamRequestContext{
		Referer:   "https://anime.example.com/watch/1",
		UserAgent: "goani-default-agent",
		Cookies:   "session=abc123",
		Headers: map[string]string{
			"Referer":    "https://cdn.example.com/override",
			"User-Agent": "source-specific-agent",
			"Cookie":     "session=override",
			"X-Trace-ID": "trace-001",
		},
	})

	if got := req.Header.Get("Referer"); got != "https://cdn.example.com/override" {
		t.Fatalf("unexpected referer: got %q", got)
	}
	if got := req.Header.Get("User-Agent"); got != "source-specific-agent" {
		t.Fatalf("unexpected user agent: got %q", got)
	}
	if got := req.Header.Get("Cookie"); got != "session=override" {
		t.Fatalf("unexpected cookie: got %q", got)
	}
	if got := req.Header.Get("X-Trace-ID"); got != "trace-001" {
		t.Fatalf("unexpected custom header: got %q", got)
	}
}

func TestBuildFFmpegHLSArgsUsesEffectiveHeaders(t *testing.T) {
	args := buildFFmpegHLSArgs(StreamRequestContext{
		SourceURL: "https://media.example.com/master.m3u8",
		Referer:   "https://anime.example.com/watch/1",
		UserAgent: "goani-default-agent",
		Cookies:   "session=abc123",
		Headers: map[string]string{
			"Referer":    "https://cdn.example.com/override",
			"User-Agent": "source-specific-agent",
			"Cookie":     "session=override",
			"X-Trace-ID": "trace-001",
		},
	})

	assertArgsContainSequence(t, args, "-user_agent", "source-specific-agent")
	assertArgsContainSequence(t, args, "-referer", "https://cdn.example.com/override")
	assertArgsContainSequence(t, args, "-cookies", "session=override")
	assertArgsContainSequence(t, args, "-i", "https://media.example.com/master.m3u8")
	assertArgsContainSequence(t, args, "-f", "mpegts")

	headersIndex := findArgIndex(args, "-headers")
	if headersIndex < 0 || headersIndex+1 >= len(args) {
		t.Fatalf("expected -headers argument, got %v", args)
	}
	if got := args[headersIndex+1]; got != "X-Trace-ID: trace-001\r\n" {
		t.Fatalf("unexpected ffmpeg headers: got %q", got)
	}
}

func assertArgsContainSequence(t *testing.T, args []string, expected ...string) {
	t.Helper()
	for i := 0; i <= len(args)-len(expected); i++ {
		match := true
		for j := range expected {
			if args[i+j] != expected[j] {
				match = false
				break
			}
		}
		if match {
			return
		}
	}
	t.Fatalf("expected args to contain %v, got %v", expected, args)
}

func findArgIndex(args []string, target string) int {
	for i, arg := range args {
		if arg == target {
			return i
		}
	}
	return -1
}
