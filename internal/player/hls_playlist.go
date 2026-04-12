package player

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	hlsm3u8 "github.com/Eyevinn/hls-m3u8/m3u8"
)

type playlistAttribute struct {
	key      string
	rawValue string
}

func decodeHLSPlaylist(body []byte) (hlsm3u8.Playlist, hlsm3u8.ListType, error) {
	var buf bytes.Buffer
	if _, err := buf.Write(body); err != nil {
		return nil, 0, err
	}
	return hlsm3u8.Decode(buf, false)
}

func looksLikeHLSPlaylist(body []byte) bool {
	for _, line := range strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		return trimmed == "#EXTM3U"
	}
	return false
}

func isHLSPlaylistBody(body []byte) bool {
	return looksLikeHLSPlaylist(body)
}

// InferM3U8Quality 尝试从 m3u8 内容中提取可展示的清晰度描述。
func InferM3U8Quality(body []byte) string {
	playlist, listType, err := decodeHLSPlaylist(body)
	if err != nil || listType != hlsm3u8.MASTER {
		return ""
	}

	master, ok := playlist.(*hlsm3u8.MasterPlaylist)
	if !ok {
		return ""
	}

	bestHeight := 0
	for _, variant := range master.Variants {
		if variant == nil {
			continue
		}

		height := parseResolutionHeight(variant.Resolution)
		if height > bestHeight {
			bestHeight = height
		}
	}
	if bestHeight == 0 {
		return ""
	}
	return fmt.Sprintf("自适应(最高%dp)", bestHeight)
}

func rewriteHLSPlaylist(target string, body []byte, host string) ([]byte, error) {
	base, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	for i, line := range lines {
		rewritten, err := rewriteHLSPlaylistLine(line, base, host)
		if err != nil {
			return nil, err
		}
		lines[i] = rewritten
	}

	return []byte(strings.Join(lines, "\n")), nil
}

func collectHLSPrefetchTargets(target string, body []byte, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}

	base, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	var initTargets []string
	var mediaTargets []string
	hasEndList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case trimmed == "#EXT-X-ENDLIST":
			hasEndList = true
		case strings.HasPrefix(trimmed, "#EXT-X-KEY:"):
			if resolved, ok := resolvePlaylistAttributeReference(trimmed, base, "URI"); ok {
				initTargets = append(initTargets, resolved)
			}
		case strings.HasPrefix(trimmed, "#EXT-X-MAP:"):
			if resolved, ok := resolvePlaylistAttributeReference(trimmed, base, "URI"); ok {
				initTargets = append(initTargets, resolved)
			}
		case strings.HasPrefix(trimmed, "#EXT-X-PART:"):
			if resolved, ok := resolvePlaylistAttributeReference(trimmed, base, "URI"); ok {
				mediaTargets = append(mediaTargets, resolved)
			}
		case strings.HasPrefix(trimmed, "#EXT-X-PRELOAD-HINT:"):
			if resolved, ok := resolvePlaylistAttributeReference(trimmed, base, "URI"); ok {
				mediaTargets = append(mediaTargets, resolved)
			}
		case strings.HasPrefix(trimmed, "#"):
			continue
		default:
			resolved, ok := resolvePlaylistLineReference(trimmed, base)
			if !ok || isLikelyPlaylistURL(resolved) {
				continue
			}
			mediaTargets = append(mediaTargets, resolved)
		}
	}

	results := uniqueStrings(initTargets)
	mediaTargets = uniqueStrings(mediaTargets)
	if len(mediaTargets) == 0 {
		return results, nil
	}

	if hasEndList {
		if len(mediaTargets) > limit {
			mediaTargets = mediaTargets[:limit]
		}
		return append(results, mediaTargets...), nil
	}

	if len(mediaTargets) > limit {
		mediaTargets = mediaTargets[len(mediaTargets)-limit:]
	}
	return append(results, mediaTargets...), nil
}

func rewriteHLSPlaylistLine(line string, base *url.URL, host string) (string, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return line, nil
	}

	if strings.HasPrefix(trimmed, "#") {
		return rewriteHLSPlaylistTagLine(line, trimmed, base, host)
	}

	rewritten, changed, err := rewritePlaylistReference(trimmed, base, host)
	if err != nil || !changed {
		return line, err
	}
	return strings.Replace(line, trimmed, rewritten, 1), nil
}

func rewriteHLSPlaylistTagLine(line, trimmed string, base *url.URL, host string) (string, error) {
	colon := strings.IndexByte(trimmed, ':')
	if colon <= 0 || colon+1 >= len(trimmed) {
		return line, nil
	}

	rawAttrs := trimmed[colon+1:]
	attrs := parsePlaylistAttributes(rawAttrs)
	if len(attrs) == 0 {
		return line, nil
	}

	changed := false
	for i := range attrs {
		if !isURIAttributeKey(attrs[i].key) {
			continue
		}

		value := strings.Trim(attrs[i].rawValue, `"`)
		rewritten, rewrittenChanged, err := rewritePlaylistReference(value, base, host)
		if err != nil {
			return "", err
		}
		if !rewrittenChanged {
			continue
		}

		attrs[i].rawValue = formatPlaylistAttributeValue(attrs[i].rawValue, rewritten)
		changed = true
	}
	if !changed {
		return line, nil
	}

	rewritten := trimmed[:colon+1] + encodePlaylistAttributes(attrs)
	return strings.Replace(line, trimmed, rewritten, 1), nil
}

func rewritePlaylistReference(raw string, base *url.URL, host string) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, nil
	}

	resolved, err := base.Parse(raw)
	if err != nil {
		return "", false, err
	}
	if !isProxyablePlaylistScheme(resolved.Scheme) {
		return raw, false, nil
	}

	return localProxyURL(host, resolved.String()), true, nil
}

func resolvePlaylistLineReference(raw string, base *url.URL) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}

	resolved, err := base.Parse(raw)
	if err != nil || !isProxyablePlaylistScheme(resolved.Scheme) {
		return "", false
	}
	return resolved.String(), true
}

func resolvePlaylistAttributeReference(line string, base *url.URL, keys ...string) (string, bool) {
	colon := strings.IndexByte(line, ':')
	if colon < 0 || colon+1 >= len(line) {
		return "", false
	}

	attrs := parsePlaylistAttributes(line[colon+1:])
	for _, attr := range attrs {
		for _, key := range keys {
			if strings.EqualFold(strings.TrimSpace(attr.key), key) {
				return resolvePlaylistLineReference(strings.Trim(attr.rawValue, `"`), base)
			}
		}
	}
	return "", false
}

func parsePlaylistAttributes(raw string) []playlistAttribute {
	if raw == "" {
		return nil
	}

	attrs := make([]playlistAttribute, 0, 8)
	i := 0
	for i < len(raw) {
		for i < len(raw) && (raw[i] == ' ' || raw[i] == ',') {
			i++
		}
		if i >= len(raw) {
			break
		}

		keyStart := i
		for i < len(raw) && raw[i] != '=' && raw[i] != ',' {
			i++
		}
		key := strings.TrimSpace(raw[keyStart:i])
		if key == "" {
			if i < len(raw) {
				i++
			}
			continue
		}
		if i >= len(raw) || raw[i] != '=' {
			continue
		}

		i++
		valueStart := i
		if i < len(raw) && raw[i] == '"' {
			i++
			for i < len(raw) {
				if raw[i] == '"' {
					i++
					break
				}
				i++
			}
		}
		for i < len(raw) && raw[i] != ',' {
			i++
		}

		value := raw[valueStart:i]
		attrs = append(attrs, playlistAttribute{
			key:      key,
			rawValue: strings.TrimSpace(value),
		})
		if i < len(raw) && raw[i] == ',' {
			i++
		}
	}

	return attrs
}

func encodePlaylistAttributes(attrs []playlistAttribute) string {
	parts := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		if attr.key == "" {
			continue
		}
		if attr.rawValue == "" {
			parts = append(parts, attr.key)
			continue
		}
		parts = append(parts, attr.key+"="+attr.rawValue)
	}
	return strings.Join(parts, ",")
}

func formatPlaylistAttributeValue(original, rewritten string) string {
	if strings.HasPrefix(original, `"`) && strings.HasSuffix(original, `"`) {
		return `"` + rewritten + `"`
	}
	return rewritten
}

func isURIAttributeKey(key string) bool {
	key = strings.ToUpper(strings.TrimSpace(key))
	return key == "URI" || strings.HasSuffix(key, "-URI") || key == "X-ASSET-LIST"
}

func isProxyablePlaylistScheme(scheme string) bool {
	scheme = strings.ToLower(strings.TrimSpace(scheme))
	return scheme == "" || scheme == "http" || scheme == "https"
}

func isLikelyPlaylistURL(target string) bool {
	parsed, err := url.Parse(target)
	if err != nil {
		return strings.Contains(strings.ToLower(target), ".m3u8")
	}
	return strings.HasSuffix(strings.ToLower(parsed.Path), ".m3u8")
}

func uniqueStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func parseResolutionHeight(resolution string) int {
	resolution = strings.TrimSpace(resolution)
	if resolution == "" {
		return 0
	}

	parts := strings.SplitN(resolution, "x", 2)
	if len(parts) != 2 {
		return 0
	}
	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return height
}
