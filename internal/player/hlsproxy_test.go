package player

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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

func TestRewritePlaylistRewritesURIAttributes(t *testing.T) {
	proxy := &hlsProxy{}
	body := []byte("#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\"key.key\"\n#EXT-X-MAP:URI=\"init.mp4\"\n#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"aac\",NAME=\"default\",URI=\"audio/index.m3u8\"\n#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=123456,URI=\"iframe.m3u8\"\n")

	rewritten, err := proxy.rewritePlaylist("https://media.example.com/path/master.m3u8", body, "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("rewritePlaylist returned error: %v", err)
	}

	got := string(rewritten)
	expectedParts := []string{
		`#EXT-X-KEY:METHOD=AES-128,URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Fpath%2Fkey.key"`,
		`#EXT-X-MAP:URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Fpath%2Finit.mp4"`,
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aac",NAME="default",URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Fpath%2Faudio%2Findex.m3u8"`,
		`#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=123456,URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Fpath%2Fiframe.m3u8"`,
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
	var gotCookie string
	var gotTraceID string
	var gotRange string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		gotReferer = r.Header.Get("Referer")
		gotCookie = r.Header.Get("Cookie")
		gotTraceID = r.Header.Get("X-Trace-ID")
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
		requestContext: StreamRequestContext{
			Referer:   "https://anime.example.com/watch/1",
			UserAgent: "goani-test-agent",
			Cookies:   "session=abc123",
			Headers: map[string]string{
				"X-Trace-ID": "trace-001",
			},
		},
		baseURL: parsedURL,
		client:  upstream.Client(),
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
	if gotCookie != "session=abc123" {
		t.Fatalf("unexpected cookie header: got %q", gotCookie)
	}
	if gotTraceID != "trace-001" {
		t.Fatalf("unexpected custom header: got %q", gotTraceID)
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

func TestServeUpstreamPreservesContentLengthForBinaryResponse(t *testing.T) {
	upstreamBody := "0123456789abcdef"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", "16")
		_, _ = io.WriteString(w, upstreamBody)
	}))
	defer upstream.Close()

	proxy := &hlsProxy{
		client: upstream.Client(),
	}

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/proxy", nil)
	rec := httptest.NewRecorder()

	proxy.serveUpstream(rec, req, upstream.URL+"/segment.ts")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Length"); got != "16" {
		t.Fatalf("unexpected content-length: got %q want %q", got, "16")
	}
	if rec.Body.String() != upstreamBody {
		t.Fatalf("unexpected body: got %q want %q", rec.Body.String(), upstreamBody)
	}
}

func TestServeUpstreamCachesSegmentResponses(t *testing.T) {
	var upstreamHits int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamHits, 1)
		w.Header().Set("Content-Type", "video/mp2t")
		_, _ = io.WriteString(w, "segment-data")
	}))
	defer upstream.Close()

	proxy := &hlsProxy{
		client: upstream.Client(),
	}

	firstReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/proxy", nil)
	firstRec := httptest.NewRecorder()
	proxy.serveUpstream(firstRec, firstReq, upstream.URL+"/segment.ts")

	secondReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/proxy", nil)
	secondRec := httptest.NewRecorder()
	proxy.serveUpstream(secondRec, secondReq, upstream.URL+"/segment.ts")

	if got := firstRec.Body.String(); got != "segment-data" {
		t.Fatalf("unexpected first response body: got %q", got)
	}
	if got := secondRec.Body.String(); got != "segment-data" {
		t.Fatalf("unexpected second response body: got %q", got)
	}
	if got := atomic.LoadInt32(&upstreamHits); got != 1 {
		t.Fatalf("expected one upstream hit after cached replay, got %d", got)
	}
}

func TestServeUpstreamPrefetchesPlaylistAssetsIntoCache(t *testing.T) {
	var playlistHits int32
	var initHits int32
	var segmentHits int32

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/media.m3u8":
			atomic.AddInt32(&playlistHits, 1)
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			_, _ = io.WriteString(w, "#EXTM3U\n#EXT-X-MAP:URI=\"init.mp4\"\n#EXTINF:1.0,\nsegment1.ts\n#EXTINF:1.0,\nsegment2.ts\n#EXT-X-ENDLIST\n")
		case "/init.mp4":
			atomic.AddInt32(&initHits, 1)
			w.Header().Set("Content-Type", "video/mp4")
			_, _ = io.WriteString(w, "init-data")
		case "/segment1.ts":
			atomic.AddInt32(&segmentHits, 1)
			w.Header().Set("Content-Type", "video/mp2t")
			_, _ = io.WriteString(w, "segment-1")
		case "/segment2.ts":
			atomic.AddInt32(&segmentHits, 1)
			w.Header().Set("Content-Type", "video/mp2t")
			_, _ = io.WriteString(w, "segment-2")
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	proxy := &hlsProxy{
		client: upstream.Client(),
	}

	playlistReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/master.m3u8", nil)
	playlistReq.Host = "127.0.0.1:9090"
	playlistRec := httptest.NewRecorder()
	proxy.serveUpstream(playlistRec, playlistReq, upstream.URL+"/media.m3u8")

	waitForCondition(t, 2*time.Second, func() bool {
		return atomic.LoadInt32(&initHits) >= 1 && atomic.LoadInt32(&segmentHits) >= 1
	})

	beforeSegmentHits := atomic.LoadInt32(&segmentHits)
	segmentReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/proxy", nil)
	segmentRec := httptest.NewRecorder()
	proxy.serveUpstream(segmentRec, segmentReq, upstream.URL+"/segment1.ts")

	if got := segmentRec.Body.String(); got != "segment-1" {
		t.Fatalf("unexpected cached segment body: got %q", got)
	}
	if got := atomic.LoadInt32(&playlistHits); got != 1 {
		t.Fatalf("unexpected playlist upstream hits: got %d", got)
	}
	if got := atomic.LoadInt32(&segmentHits); got != beforeSegmentHits {
		t.Fatalf("expected prefetched segment to be served from cache, got upstream hits %d -> %d", beforeSegmentHits, got)
	}
}

func TestPrefetchAssetSkipsAlreadyCachedResponse(t *testing.T) {
	var upstreamHits int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamHits, 1)
		w.Header().Set("Content-Type", "video/mp2t")
		_, _ = io.WriteString(w, "segment-data")
	}))
	defer upstream.Close()

	proxy := &hlsProxy{
		client: upstream.Client(),
	}

	proxy.prefetchAsset(upstream.URL + "/segment.ts")
	proxy.prefetchAsset(upstream.URL + "/segment.ts")

	if got := atomic.LoadInt32(&upstreamHits); got != 1 {
		t.Fatalf("expected cached prefetch to avoid duplicate upstream fetch, got %d", got)
	}
	if cached := proxy.loadCacheEntry(upstream.URL+"/segment.ts", ""); cached == nil {
		t.Fatalf("expected prefetched segment to be cached")
	}
}

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}
