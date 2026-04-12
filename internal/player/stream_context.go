package player

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// StreamRequestContext 描述访问播放链接时需要携带的请求上下文。
// 它会被 HLS 代理和 ffmpeg 桥接共同复用，避免不同播放链路各自维护一套 header 逻辑。
type StreamRequestContext struct {
	SourceURL string            `json:"sourceUrl"`
	Referer   string            `json:"referer,omitempty"`
	UserAgent string            `json:"userAgent,omitempty"`
	Cookies   string            `json:"cookies,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type headerPair struct {
	key   string
	value string
}

// Normalized 返回一个去掉空白和空 header 的上下文副本。
func (c StreamRequestContext) Normalized() StreamRequestContext {
	normalized := StreamRequestContext{
		SourceURL: strings.TrimSpace(c.SourceURL),
		Referer:   strings.TrimSpace(c.Referer),
		UserAgent: strings.TrimSpace(c.UserAgent),
		Cookies:   strings.TrimSpace(c.Cookies),
	}

	if len(c.Headers) == 0 {
		return normalized
	}

	headers := make(map[string]string, len(c.Headers))
	for key, value := range c.Headers {
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			continue
		}
		headers[trimmedKey] = trimmedValue
	}
	if len(headers) > 0 {
		normalized.Headers = headers
	}
	return normalized
}

// EncodeStreamRequestContext 把请求上下文编码成适合命令行参数传递的字符串。
func EncodeStreamRequestContext(ctx StreamRequestContext) (string, error) {
	payload, err := json.Marshal(ctx.Normalized())
	if err != nil {
		return "", fmt.Errorf("序列化请求上下文失败: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

// DecodeStreamRequestContext 从命令行参数恢复请求上下文。
func DecodeStreamRequestContext(encoded string) (StreamRequestContext, error) {
	data, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return StreamRequestContext{}, fmt.Errorf("解码请求上下文失败: %w", err)
	}

	var ctx StreamRequestContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return StreamRequestContext{}, fmt.Errorf("解析请求上下文失败: %w", err)
	}
	return ctx.Normalized(), nil
}

func (c StreamRequestContext) effectiveHeaders() (string, string, string, []headerPair) {
	normalized := c.Normalized()
	userAgent := normalized.UserAgent
	referer := normalized.Referer
	cookies := normalized.Cookies

	if len(normalized.Headers) == 0 {
		return userAgent, referer, cookies, nil
	}

	keys := make([]string, 0, len(normalized.Headers))
	for key := range normalized.Headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	extras := make([]headerPair, 0, len(keys))
	for _, key := range keys {
		value := normalized.Headers[key]
		switch strings.ToLower(key) {
		case "user-agent":
			userAgent = value
		case "referer":
			referer = value
		case "cookie":
			cookies = value
		default:
			extras = append(extras, headerPair{key: key, value: value})
		}
	}

	return userAgent, referer, cookies, extras
}

func applyStreamRequestContext(req *http.Request, ctx StreamRequestContext) {
	userAgent, referer, cookies, extras := ctx.effectiveHeaders()

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}
	for _, header := range extras {
		req.Header.Set(header.key, header.value)
	}
}

// MPVHTTPArgs 把请求上下文转换成 mpv 可识别的网络参数。
func (c StreamRequestContext) MPVHTTPArgs() []string {
	userAgent, referer, cookies, extras := c.effectiveHeaders()
	args := make([]string, 0, 4+len(extras))

	if userAgent != "" {
		args = append(args, "--user-agent="+userAgent)
	}
	if referer != "" {
		args = append(args, "--referrer="+referer)
	}
	if cookies != "" {
		args = append(args, "--http-header-fields-append=Cookie: "+cookies)
	}
	for _, header := range extras {
		args = append(args, "--http-header-fields-append="+header.key+": "+header.value)
	}

	return args
}
