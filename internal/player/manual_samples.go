package player

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PlaybackSample 表示一条用于本地播放验证的真实样本。
// 它通常来自一次现场采样，包含视频地址和访问它所需的请求上下文。
type PlaybackSample struct {
	CollectedAt string            `json:"collectedAt,omitempty"`
	Keyword     string            `json:"keyword,omitempty"`
	AnimeName   string            `json:"animeName,omitempty"`
	EpisodeName string            `json:"episodeName,omitempty"`
	SourceID    string            `json:"sourceId,omitempty"`
	SourceName  string            `json:"sourceName,omitempty"`
	EpisodeURL  string            `json:"episodeUrl,omitempty"`
	VideoURL    string            `json:"videoUrl,omitempty"`
	Cookies     string            `json:"cookies,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// StreamRequestContext 将样本转换为播放器链可复用的请求上下文。
func (s PlaybackSample) StreamRequestContext(defaultUserAgent string) StreamRequestContext {
	return StreamRequestContext{
		SourceURL: s.VideoURL,
		Referer:   s.EpisodeURL,
		UserAgent: defaultUserAgent,
		Cookies:   s.Cookies,
		Headers:   cloneStringMap(s.Headers),
	}
}

// PlaybackSampleFile 表示一次采样产出的完整样本文件。
type PlaybackSampleFile struct {
	CollectedAt string           `json:"collectedAt,omitempty"`
	Generator   string           `json:"generator,omitempty"`
	Cases       []PlaybackSample `json:"cases"`
}

// SavePlaybackSampleFile 保存样本文件。
func SavePlaybackSampleFile(path string, file PlaybackSampleFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadPlaybackSampleFile 读取样本文件。
func LoadPlaybackSampleFile(path string) (*PlaybackSampleFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file PlaybackSampleFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
