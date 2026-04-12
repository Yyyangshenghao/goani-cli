package player

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRewritePlaylistRewritesLowLatencyAndExtendedURIAttributes(t *testing.T) {
	proxy := &hlsProxy{}
	body := []byte("#EXTM3U\n" +
		"#EXT-X-PART:DURATION=0.334,URI=\"parts/part-1.m4s\"\n" +
		"#EXT-X-PRELOAD-HINT:TYPE=PART,URI=\"parts/part-2.m4s\"\n" +
		"#EXT-X-RENDITION-REPORT:URI=\"renditions/audio.m3u8\",LAST-MSN=15,LAST-PART=2\n" +
		"#EXT-X-CONTENT-STEERING:SERVER-URI=\"steering.json\",PATHWAY-ID=\"cdn-a\"\n" +
		"#EXT-X-DATERANGE:ID=\"ad-1\",CLASS=\"com.apple.hls.preload\",START-DATE=\"2020-01-02T21:55:44Z\",DURATION=15.000,X-URI=\"ads/ad.m3u8\",X-ASSET-URI=\"ads/asset.m3u8\",X-ASSET-LIST=\"ads/list.json\"\n")

	rewritten, err := proxy.rewritePlaylist("https://media.example.com/live/master.m3u8", body, "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("rewritePlaylist returned error: %v", err)
	}

	got := string(rewritten)
	expectedParts := []string{
		`#EXT-X-PART:DURATION=0.334,URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Flive%2Fparts%2Fpart-1.m4s"`,
		`#EXT-X-PRELOAD-HINT:TYPE=PART,URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Flive%2Fparts%2Fpart-2.m4s"`,
		`#EXT-X-RENDITION-REPORT:URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Flive%2Frenditions%2Faudio.m3u8",LAST-MSN=15,LAST-PART=2`,
		`#EXT-X-CONTENT-STEERING:SERVER-URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Flive%2Fsteering.json",PATHWAY-ID="cdn-a"`,
		`X-URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Flive%2Fads%2Fad.m3u8"`,
		`X-ASSET-URI="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Flive%2Fads%2Fasset.m3u8"`,
		`X-ASSET-LIST="http://127.0.0.1:8080/proxy?u=https%3A%2F%2Fmedia.example.com%2Flive%2Fads%2Flist.json"`,
	}
	for _, part := range expectedParts {
		if !strings.Contains(got, part) {
			t.Fatalf("rewritten playlist missing %q, got:\n%s", part, got)
		}
	}
}

func TestRewritePlaylistLeavesUnsupportedSchemesUntouched(t *testing.T) {
	proxy := &hlsProxy{}
	body := []byte("#EXTM3U\n#EXT-X-KEY:METHOD=SAMPLE-AES,URI=\"skd://fairplay-key\"\n")

	rewritten, err := proxy.rewritePlaylist("https://media.example.com/path/master.m3u8", body, "127.0.0.1:8080")
	if err != nil {
		t.Fatalf("rewritePlaylist returned error: %v", err)
	}

	got := string(rewritten)
	if !strings.Contains(got, `URI="skd://fairplay-key"`) {
		t.Fatalf("expected unsupported scheme to remain unchanged, got:\n%s", got)
	}
	if strings.Contains(got, "/proxy?u=skd%3A%2F%2Ffairplay-key") {
		t.Fatalf("unsupported scheme should not be proxied, got:\n%s", got)
	}
}

func TestInferM3U8QualityUsesStructuredMasterPlaylistParsing(t *testing.T) {
	body := []byte("#EXTM3U\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720\n720p/index.m3u8\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1920x1080\n1080p/index.m3u8\n")

	got := InferM3U8Quality(body)
	if got != "自适应(最高1080p)" {
		t.Fatalf("unexpected quality: got %q", got)
	}
}

func TestServeUpstreamRewritesTextPlainPlaylistWithoutM3U8Suffix(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "#EXTM3U\nchunk.ts\n")
	}))
	defer upstream.Close()

	proxy := &hlsProxy{
		client: upstream.Client(),
	}

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/proxy", nil)
	req.Host = "127.0.0.1:9090"
	rec := httptest.NewRecorder()

	proxy.serveUpstream(rec, req, upstream.URL+"/playlist?id=1")

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/vnd.apple.mpegurl" {
		t.Fatalf("unexpected content-type: got %q", got)
	}
	if !strings.Contains(rec.Body.String(), "http://127.0.0.1:9090/proxy?u=") {
		t.Fatalf("expected playlist body to be rewritten, got:\n%s", rec.Body.String())
	}
}

func TestCollectHLSPrefetchTargetsPrefersVODStart(t *testing.T) {
	body := []byte("#EXTM3U\n#EXT-X-MAP:URI=\"init.mp4\"\n#EXTINF:1.0,\nsegment1.ts\n#EXTINF:1.0,\nsegment2.ts\n#EXT-X-ENDLIST\n")

	targets, err := collectHLSPrefetchTargets("https://media.example.com/vod/index.m3u8", body, 2)
	if err != nil {
		t.Fatalf("collectHLSPrefetchTargets returned error: %v", err)
	}

	want := []string{
		"https://media.example.com/vod/init.mp4",
		"https://media.example.com/vod/segment1.ts",
		"https://media.example.com/vod/segment2.ts",
	}
	if strings.Join(targets, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected targets:\n%s", strings.Join(targets, "\n"))
	}
}

func TestCollectHLSPrefetchTargetsPrefersLiveTail(t *testing.T) {
	body := []byte("#EXTM3U\n#EXT-X-MAP:URI=\"init.mp4\"\n#EXTINF:1.0,\nsegment1.ts\n#EXTINF:1.0,\nsegment2.ts\n#EXT-X-PART:DURATION=0.3,URI=\"part1.m4s\"\n#EXT-X-PRELOAD-HINT:TYPE=PART,URI=\"part2.m4s\"\n")

	targets, err := collectHLSPrefetchTargets("https://media.example.com/live/index.m3u8", body, 2)
	if err != nil {
		t.Fatalf("collectHLSPrefetchTargets returned error: %v", err)
	}

	want := []string{
		"https://media.example.com/live/init.mp4",
		"https://media.example.com/live/part1.m4s",
		"https://media.example.com/live/part2.m4s",
	}
	if strings.Join(targets, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected targets:\n%s", strings.Join(targets, "\n"))
	}
}
