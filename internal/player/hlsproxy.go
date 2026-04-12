package player

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

const proxyIdleTimeout = 30 * time.Minute

// StartDetachedHLSProxy 启动后台 HLS 代理，并返回可交给播放器的本地 m3u8 地址。
func StartDetachedHLSProxy(ctx StreamRequestContext) (string, error) {
	return startDetachedHelper("proxy-hls", ctx)
}

// ServeHLSProxy 运行本地 HLS 代理服务。
// 这个代理的目标不是转码，而是把远端需要请求头的 m3u8/ts 请求，
// 包装成本地播放器更容易消费的一条本地地址。
func ServeHLSProxy(ctx StreamRequestContext, out io.Writer) error {
	ctx = ctx.Normalized()
	if ctx.SourceURL == "" {
		return fmt.Errorf("缺少播放地址")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()

	baseURL, err := url.Parse(ctx.SourceURL)
	if err != nil {
		return fmt.Errorf("无效的源地址: %w", err)
	}

	proxy := &hlsProxy{
		requestContext: ctx,
		baseURL:        baseURL,
		client:         &http.Client{Timeout: 20 * time.Second},
		lastAccess:     time.Now().Unix(),
	}

	server := &http.Server{
		Handler: proxy.routes(),
	}

	go proxy.shutdownWhenIdle(server)

	localURL := fmt.Sprintf("http://%s/master.m3u8", listener.Addr().String())
	if _, err := fmt.Fprintln(out, localURL); err != nil {
		return err
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

type hlsProxy struct {
	requestContext StreamRequestContext
	baseURL        *url.URL
	client         *http.Client
	lastAccess     int64
}

// routes 暴露两类入口：
// 1. /master.m3u8 作为播放器看到的主入口
// 2. /proxy      代理后续子 playlist 和分片请求
func (p *hlsProxy) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/master.m3u8", func(w http.ResponseWriter, r *http.Request) {
		p.touch()
		p.serveUpstream(w, r, p.requestContext.SourceURL)
	})
	mux.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		p.touch()
		target := strings.TrimSpace(r.URL.Query().Get("u"))
		if target == "" {
			http.Error(w, "missing upstream url", http.StatusBadRequest)
			return
		}
		p.serveUpstream(w, r, target)
	})
	return mux
}

// serveUpstream 转发单个上游请求，并在遇到 m3u8 时改写其中的链接，
// 让后续所有子请求继续回到本地代理，而不是直接打到远端站点。
func (p *hlsProxy) serveUpstream(w http.ResponseWriter, incoming *http.Request, target string) {
	req, err := http.NewRequestWithContext(incoming.Context(), http.MethodGet, target, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	applyStreamRequestContext(req, p.requestContext)
	if rangeHeader := strings.TrimSpace(incoming.Header.Get("Range")); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "mpegurl") || strings.HasSuffix(strings.ToLower(target), ".m3u8") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		rewritten, err := p.rewritePlaylist(target, body, incoming.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, bytes.NewReader(rewritten))
		return
	}

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// rewritePlaylist 会把 playlist 里的相对/绝对资源地址统一改写成本地代理地址，
// 这样播放器后续访问子 m3u8、ts 分片时，仍然会带上 goani 负责维护的请求上下文。
func (p *hlsProxy) rewritePlaylist(target string, body []byte, host string) ([]byte, error) {
	base, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		rewrittenTagLine, changed, err := rewritePlaylistTagURIs(line, base, host)
		if err != nil {
			return nil, err
		}
		if changed {
			lines[i] = rewrittenTagLine
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		resolved, err := base.Parse(trimmed)
		if err != nil {
			continue
		}

		lines[i] = localProxyURL(host, resolved.String())
	}

	return []byte(strings.Join(lines, "\n")), nil
}

var playlistURIPattern = regexp.MustCompile(`URI="([^"]+)"`)

func rewritePlaylistTagURIs(line string, base *url.URL, host string) (string, bool, error) {
	if !strings.Contains(line, `URI="`) {
		return line, false, nil
	}

	var rewriteErr error
	changed := false
	rewritten := playlistURIPattern.ReplaceAllStringFunc(line, func(match string) string {
		if rewriteErr != nil {
			return match
		}

		submatches := playlistURIPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		resolved, err := base.Parse(strings.TrimSpace(submatches[1]))
		if err != nil {
			rewriteErr = err
			return match
		}

		changed = true
		return fmt.Sprintf(`URI="%s"`, localProxyURL(host, resolved.String()))
	})
	if rewriteErr != nil {
		return "", false, rewriteErr
	}

	return rewritten, changed, nil
}

func localProxyURL(host, target string) string {
	return fmt.Sprintf("http://%s/proxy?u=%s", host, url.QueryEscape(target))
}

func (p *hlsProxy) touch() {
	atomic.StoreInt64(&p.lastAccess, time.Now().Unix())
}

// shutdownWhenIdle 在长时间没有播放器访问时自动关闭代理，
// 避免每次播放后留下常驻本地端口进程。
func (p *hlsProxy) shutdownWhenIdle(server *http.Server) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		last := time.Unix(atomic.LoadInt64(&p.lastAccess), 0)
		if time.Since(last) < proxyIdleTimeout {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = server.Shutdown(ctx)
		cancel()
		return
	}
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
