package player

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

var ffmpegLookPath = exec.LookPath

// StartDetachedFFmpegHLSBridge 启动后台 ffmpeg 桥接，并返回本地可播放的流地址。
func StartDetachedFFmpegHLSBridge(ctx StreamRequestContext) (string, error) {
	if _, err := ffmpegLookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("未找到 ffmpeg: %w", err)
	}
	return startDetachedHelper("bridge-hls", ctx)
}

// ServeFFmpegHLSBridge 运行本地 ffmpeg HLS 桥接服务。
// 它会把需要特殊请求头的 m3u8 拉流并转成更容易被播放器直接消费的 mpegts。
func ServeFFmpegHLSBridge(ctx StreamRequestContext, out io.Writer) error {
	ctx = ctx.Normalized()
	if ctx.SourceURL == "" {
		return fmt.Errorf("缺少播放地址")
	}

	ffmpegPath, err := ffmpegLookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("未找到 ffmpeg: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()

	bridge := &ffmpegHLSBridge{
		ffmpegPath:     ffmpegPath,
		requestContext: ctx,
		lastAccess:     time.Now().Unix(),
	}

	server := &http.Server{
		Handler: bridge.routes(),
	}

	go bridge.shutdownWhenIdle(server)

	localURL := fmt.Sprintf("http://%s/live.ts", listener.Addr().String())
	if _, err := fmt.Fprintln(out, localURL); err != nil {
		return err
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

type ffmpegHLSBridge struct {
	ffmpegPath     string
	requestContext StreamRequestContext
	lastAccess     int64
}

func (b *ffmpegHLSBridge) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/live.ts", b.serveStream)
	return mux
}

func (b *ffmpegHLSBridge) serveStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	b.touch()

	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	cmd := exec.CommandContext(r.Context(), b.ffmpegPath, buildFFmpegHLSArgs(b.requestContext)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	_, copyErr := io.Copy(w, stdout)
	waitErr := cmd.Wait()
	if copyErr != nil && !isStreamTerminationError(copyErr) {
		return
	}
	if waitErr != nil && !isStreamTerminationError(waitErr) {
		return
	}
}

func (b *ffmpegHLSBridge) touch() {
	atomic.StoreInt64(&b.lastAccess, time.Now().Unix())
}

func (b *ffmpegHLSBridge) shutdownWhenIdle(server *http.Server) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		last := time.Unix(atomic.LoadInt64(&b.lastAccess), 0)
		if time.Since(last) < proxyIdleTimeout {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = server.Shutdown(ctx)
		cancel()
		return
	}
}

func buildFFmpegHLSArgs(ctx StreamRequestContext) []string {
	args := []string{
		"-nostdin",
		"-loglevel", "error",
	}

	userAgent, referer, cookies, extraHeaders := ctx.effectiveHeaders()
	if userAgent != "" {
		args = append(args, "-user_agent", userAgent)
	}
	if referer != "" {
		args = append(args, "-referer", referer)
	}
	if cookies != "" {
		args = append(args, "-cookies", cookies)
	}
	if headerLines := formatFFmpegHeaders(extraHeaders); headerLines != "" {
		args = append(args, "-headers", headerLines)
	}

	args = append(args,
		"-i", ctx.Normalized().SourceURL,
		"-c", "copy",
		"-f", "mpegts",
		"pipe:1",
	)
	return args
}

func formatFFmpegHeaders(headers []headerPair) string {
	if len(headers) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, header := range headers {
		key := sanitizeHeaderToken(header.key)
		value := sanitizeHeaderToken(header.value)
		if key == "" || value == "" {
			continue
		}
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(value)
		builder.WriteString("\r\n")
	}
	return builder.String()
}

func sanitizeHeaderToken(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return strings.TrimSpace(value)
}

func isStreamTerminationError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "broken pipe") || strings.Contains(message, "context canceled")
}
