package player

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	proxyIdleTimeout      = 30 * time.Minute
	hlsSegmentCacheTTL    = 2 * time.Minute
	hlsKeyCacheTTL        = 10 * time.Minute
	hlsPrefetchTimeout    = 12 * time.Second
	hlsPrefetchAssetLimit = 4
	hlsCacheMaxEntries    = 96
	hlsCacheMaxEntryBytes = 8 << 20
	hlsCacheMaxBytes      = 64 << 20
)

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
		client:         newHLSProxyHTTPClient(),
		cacheEntries:   make(map[string]*proxyCacheEntry),
		inflightFetch:  make(map[string]*proxyInflightFetch),
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
	cacheMu        sync.Mutex
	cacheEntries   map[string]*proxyCacheEntry
	inflightFetch  map[string]*proxyInflightFetch
	cacheBytes     int
}

type proxyCacheEntry struct {
	response   *proxyUpstreamResponse
	expiresAt  time.Time
	lastAccess time.Time
}

type proxyInflightFetch struct {
	done     chan struct{}
	response *proxyUpstreamResponse
	err      error
}

type proxyUpstreamResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

func newHLSProxyHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 32
	transport.MaxIdleConnsPerHost = 16
	transport.MaxConnsPerHost = 32
	transport.IdleConnTimeout = 90 * time.Second

	return &http.Client{
		Timeout:   20 * time.Second,
		Transport: transport,
	}
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
	rangeHeader := strings.TrimSpace(incoming.Header.Get("Range"))
	if cached := p.loadCacheEntry(target, rangeHeader); cached != nil {
		p.writeCachedResponse(w, incoming.Method, cached)
		return
	}

	var inflight *proxyInflightFetch
	if rangeHeader == "" && incoming.Method == http.MethodGet {
		waitingFetch, owner := p.beginInflightFetch(target)
		if !owner {
			if cached := p.waitForInflightFetch(incoming.Context(), target, waitingFetch); cached != nil {
				p.writeCachedResponse(w, incoming.Method, cached)
				return
			}
		} else {
			inflight = waitingFetch
		}
	}

	req, err := http.NewRequestWithContext(incoming.Context(), http.MethodGet, target, nil)
	if err != nil {
		p.finishInflightFetch(target, inflight, nil, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	applyStreamRequestContext(req, p.requestContext)
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		p.finishInflightFetch(target, inflight, nil, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if shouldInspectAsHLSPlaylist(contentType, target) {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			p.finishInflightFetch(target, inflight, nil, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		if isHLSPlaylistResponse(target, contentType, body) {
			rewritten, err := p.rewritePlaylist(target, body, incoming.Host)
			if err != nil {
				p.finishInflightFetch(target, inflight, nil, err)
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}

			copyHeaders(w.Header(), resp.Header)
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Del("Content-Length")
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, bytes.NewReader(rewritten))
			p.prefetchPlaylistAssets(target, body)
			p.finishInflightFetch(target, inflight, nil, nil)
			return
		}

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, bytes.NewReader(body))
		p.finishInflightFetch(target, inflight, nil, nil)
		return
	}

	cacheable := inflight != nil && isCacheableBinaryResponse(target, contentType, resp.StatusCode)
	copyHeaders(w.Header(), resp.Header)
	w.Header().Del("Transfer-Encoding")
	w.WriteHeader(resp.StatusCode)
	if incoming.Method == http.MethodHead {
		p.finishInflightFetch(target, inflight, nil, nil)
		return
	}

	var buffer bytes.Buffer
	writer := io.Writer(w)
	if cacheable {
		writer = io.MultiWriter(w, &buffer)
	}

	_, copyErr := io.Copy(writer, resp.Body)
	var cachedResponse *proxyUpstreamResponse
	if cacheable && copyErr == nil {
		cachedResponse = &proxyUpstreamResponse{
			statusCode: resp.StatusCode,
			header:     resp.Header.Clone(),
			body:       append([]byte(nil), buffer.Bytes()...),
		}
		if !p.storeCacheEntry(target, cachedResponse, cacheTTLForAsset(target, contentType)) {
			cachedResponse = nil
		}
	}

	p.finishInflightFetch(target, inflight, cachedResponse, copyErr)
}

// rewritePlaylist 会把 playlist 里的相对/绝对资源地址统一改写成本地代理地址，
// 这样播放器后续访问子 m3u8、ts 分片时，仍然会带上 goani 负责维护的请求上下文。
func (p *hlsProxy) rewritePlaylist(target string, body []byte, host string) ([]byte, error) {
	return rewriteHLSPlaylist(target, body, host)
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

func shouldInspectAsHLSPlaylist(contentType, target string) bool {
	if strings.Contains(contentType, "mpegurl") || strings.HasSuffix(strings.ToLower(target), ".m3u8") {
		return true
	}
	return strings.HasPrefix(contentType, "text/")
}

func isHLSPlaylistResponse(target, contentType string, body []byte) bool {
	if strings.Contains(contentType, "mpegurl") || strings.HasSuffix(strings.ToLower(target), ".m3u8") {
		return true
	}
	return isHLSPlaylistBody(body)
}

func (p *hlsProxy) writeCachedResponse(w http.ResponseWriter, method string, response *proxyUpstreamResponse) {
	if response == nil {
		return
	}

	copyHeaders(w.Header(), response.header)
	w.Header().Del("Transfer-Encoding")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response.body)))
	w.WriteHeader(response.statusCode)
	if method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, bytes.NewReader(response.body))
}

func (p *hlsProxy) loadCacheEntry(target, rangeHeader string) *proxyUpstreamResponse {
	if rangeHeader != "" {
		return nil
	}

	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.ensureCacheStateLocked()

	entry, ok := p.cacheEntries[target]
	if !ok {
		return nil
	}
	if time.Now().After(entry.expiresAt) {
		p.removeCacheEntryLocked(target)
		return nil
	}

	entry.lastAccess = time.Now()
	return entry.response
}

func (p *hlsProxy) storeCacheEntry(target string, response *proxyUpstreamResponse, ttl time.Duration) bool {
	if response == nil || ttl <= 0 {
		return false
	}
	if len(response.body) == 0 || len(response.body) > hlsCacheMaxEntryBytes {
		return false
	}

	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.ensureCacheStateLocked()

	if existing, ok := p.cacheEntries[target]; ok {
		p.cacheBytes -= len(existing.response.body)
	}

	p.cacheEntries[target] = &proxyCacheEntry{
		response:   response,
		expiresAt:  time.Now().Add(ttl),
		lastAccess: time.Now(),
	}
	p.cacheBytes += len(response.body)
	p.pruneCacheLocked()
	return true
}

func (p *hlsProxy) beginInflightFetch(target string) (*proxyInflightFetch, bool) {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.ensureCacheStateLocked()

	if fetch, ok := p.inflightFetch[target]; ok {
		return fetch, false
	}

	fetch := &proxyInflightFetch{done: make(chan struct{})}
	p.inflightFetch[target] = fetch
	return fetch, true
}

func (p *hlsProxy) waitForInflightFetch(ctx context.Context, target string, fetch *proxyInflightFetch) *proxyUpstreamResponse {
	if fetch == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return nil
	case <-fetch.done:
		if fetch.response != nil {
			return fetch.response
		}
		return p.loadCacheEntry(target, "")
	}
}

func (p *hlsProxy) finishInflightFetch(target string, fetch *proxyInflightFetch, response *proxyUpstreamResponse, err error) {
	if fetch == nil {
		return
	}

	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.ensureCacheStateLocked()

	delete(p.inflightFetch, target)
	fetch.response = response
	fetch.err = err
	close(fetch.done)
}

func (p *hlsProxy) ensureCacheStateLocked() {
	if p.cacheEntries == nil {
		p.cacheEntries = make(map[string]*proxyCacheEntry)
	}
	if p.inflightFetch == nil {
		p.inflightFetch = make(map[string]*proxyInflightFetch)
	}
}

func (p *hlsProxy) pruneCacheLocked() {
	now := time.Now()
	for target, entry := range p.cacheEntries {
		if now.After(entry.expiresAt) {
			p.removeCacheEntryLocked(target)
		}
	}

	for len(p.cacheEntries) > hlsCacheMaxEntries || p.cacheBytes > hlsCacheMaxBytes {
		oldestTarget := ""
		var oldestTime time.Time
		for target, entry := range p.cacheEntries {
			if oldestTarget == "" || entry.lastAccess.Before(oldestTime) {
				oldestTarget = target
				oldestTime = entry.lastAccess
			}
		}
		if oldestTarget == "" {
			return
		}
		p.removeCacheEntryLocked(oldestTarget)
	}
}

func (p *hlsProxy) removeCacheEntryLocked(target string) {
	entry, ok := p.cacheEntries[target]
	if !ok {
		return
	}
	p.cacheBytes -= len(entry.response.body)
	delete(p.cacheEntries, target)
}

func (p *hlsProxy) prefetchPlaylistAssets(target string, body []byte) {
	targets, err := collectHLSPrefetchTargets(target, body, hlsPrefetchAssetLimit)
	if err != nil || len(targets) == 0 {
		return
	}

	go func() {
		for _, assetTarget := range targets {
			p.prefetchAsset(assetTarget)
		}
	}()
}

func (p *hlsProxy) prefetchAsset(target string) {
	if p.loadCacheEntry(target, "") != nil {
		return
	}

	fetch, owner := p.beginInflightFetch(target)
	if !owner {
		return
	}
	if cached := p.loadCacheEntry(target, ""); cached != nil {
		p.finishInflightFetch(target, fetch, cached, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), hlsPrefetchTimeout)
	defer cancel()

	response, err := p.fetchBinaryAsset(ctx, target)
	if err == nil && response != nil {
		contentType := strings.ToLower(response.header.Get("Content-Type"))
		if !p.storeCacheEntry(target, response, cacheTTLForAsset(target, contentType)) {
			response = nil
		}
	}

	p.finishInflightFetch(target, fetch, response, err)
}

func (p *hlsProxy) fetchBinaryAsset(ctx context.Context, target string) (*proxyUpstreamResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	applyStreamRequestContext(req, p.requestContext)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !isCacheableBinaryResponse(target, contentType, resp.StatusCode) {
		return nil, fmt.Errorf("uncacheable prefetch response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &proxyUpstreamResponse{
		statusCode: resp.StatusCode,
		header:     resp.Header.Clone(),
		body:       body,
	}, nil
}

func isCacheableBinaryResponse(target, contentType string, statusCode int) bool {
	if statusCode != http.StatusOK {
		return false
	}
	return cacheTTLForAsset(target, contentType) > 0
}

func cacheTTLForAsset(target, contentType string) time.Duration {
	switch classifyAssetKind(target, contentType) {
	case "key":
		return hlsKeyCacheTTL
	case "media":
		return hlsSegmentCacheTTL
	default:
		return 0
	}
}

func classifyAssetKind(target, contentType string) string {
	ext := strings.ToLower(filepath.Ext(pathFromTarget(target)))
	switch ext {
	case ".key":
		return "key"
	case ".ts", ".m4s", ".m4a", ".m4v", ".mp4", ".aac", ".mp3", ".cmfa", ".cmfv", ".vtt", ".webvtt":
		return "media"
	}

	switch {
	case strings.HasPrefix(contentType, "video/"), strings.HasPrefix(contentType, "audio/"):
		return "media"
	case strings.Contains(contentType, "mp4"), strings.Contains(contentType, "mpeg"), strings.Contains(contentType, "octet-stream"), strings.Contains(contentType, "webvtt"):
		return "media"
	default:
		return ""
	}
}

func pathFromTarget(target string) string {
	parsed, err := url.Parse(target)
	if err != nil {
		return target
	}
	return parsed.Path
}
