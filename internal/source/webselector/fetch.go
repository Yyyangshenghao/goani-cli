package webselector

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// fetchHTML 获取 HTML 文档
func (s *WebSelectorSource) fetchHTML(url string) (*goquery.Document, error) {
	body, err := s.fetchBytes(url)
	if err != nil {
		return nil, err
	}

	// 尝试解码 JSON 字符串
	var htmlStr string
	if err := json.Unmarshal(body, &htmlStr); err == nil {
		body = []byte(htmlStr)
	}

	return goquery.NewDocumentFromReader(strings.NewReader(string(body)))
}

// fetchBytes 获取原始字节
func (s *WebSelectorSource) fetchBytes(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// fetchRawHTML 获取原始 HTML 字符串
func (s *WebSelectorSource) fetchRawHTML(url string) (string, error) {
	body, err := s.fetchBytes(url)
	if err != nil {
		return "", err
	}

	// 尝试解码 JSON 字符串
	var htmlStr string
	if err := json.Unmarshal(body, &htmlStr); err == nil {
		return htmlStr, nil
	}
	return string(body), nil
}

// FetchRawHTML 公开的方法，用于调试
func (s *WebSelectorSource) FetchRawHTML(url string) (string, error) {
	return s.fetchRawHTML(url)
}
