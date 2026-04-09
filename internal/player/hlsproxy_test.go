package player

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRewritePlaylistRewritesRelativeAndAbsoluteLinks(t *testing.T) {
	proxy := &hlsProxy{}
	body := []byte("#EXTM3U\nsegment1.ts\n/sub/segment2.ts\nhttps://cdn.example.com/segment3.ts\n")

	rewritten, err := proxy.rewritePlaylist("https://media.example.com/path/master.m3u8", body, "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("rewritePlaylist returned error: %v", err)
	}

	got := string(rewritten)
	expectedParts := []string{
		"#EXTM3U",
		"http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Fpath%2Fsegment1.ts",
		"http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Fsub%2Fsegment2.ts",
		"http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fcdn.example.com%2Fsegment3.ts",
	}
	for _, part := range expectedParts {
		if !strings.Contains(got, part) {
			t.Fatalf("rewritten playlist missing %q, got:\n%s", part, got)
		}
	}
}

func TestServeUpstreamForwardsHeadersAndRewritesM3U8(t *testing.T) {
	var gotUserAgent string
	var gotReferer string
	var gotRange string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		gotReferer = r.Header.Get("Referer")
		gotRange = r.Header.Get("Range")
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = io.WriteString(w, "#EXTM3U\nchunk.ts\n")
	}))
	defer upstream.Close()

	parsedURL, err := url.Parse(upstream.URL + "/master.m3u8")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	proxy := &hlsProxy{
		referer:   "https://anime.example.com/watch/1",
		userAgent: "goani-test-agent",
		baseURL:   parsedURL,
		client:    upstream.Client(),
	}

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/master.m3u8", nil)
	req.Host = "127.0.0.1:9090"
	req.Header.Set("Range", "bytes=0-10")
	rec := httptest.NewRecorder()

	proxy.serveUpstream(rec, req, upstream.URL+"/master.m3u8")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}
	if gotUserAgent != "goani-test-agent" {
		t.Fatalf("unexpected user-agent: got %q", gotUserAgent)
	}
	if gotReferer != "https://anime.example.com/watch/1" {
		t.Fatalf("unexpected referer: got %q", gotReferer)
	}
	if gotRange != "bytes=0-10" {
		t.Fatalf("unexpected range header: got %q", gotRange)
	}

	body := rec.Body.String()
	expectedChunk := "http://127.0.0.1:9090/proxy?u=" + url.QueryEscape(upstream.URL+"/chunk.ts")
	if !strings.Contains(body, expectedChunk) {
		t.Fatalf("rewritten playlist missing chunk url %q, got:\n%s", expectedChunk, body)
	}
}
