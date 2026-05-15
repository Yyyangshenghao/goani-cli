package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/Yyyangshenghao/goani-cli/internal/app"
	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
	"github.com/Yyyangshenghao/goani-cli/internal/version"
)

// NewServer 创建 MCP server 并注册所有 tools
func NewServer() *mcpgo.Server {
	s := mcpgo.NewServer(
		&mcpgo.Implementation{
			Name:    "goani",
			Version: version.Version,
		},
		nil,
	)

	registerSearchTool(s)
	registerGetEpisodesTool(s)
	registerGetVideoURLTool(s)
	registerListSourcesTool(s)

	return s
}

// --- search tool ---

type searchInput struct {
	Keyword string `json:"keyword" jsonschema:"搜索关键词" jsonschema_required:"true"`
}

type searchOutput struct {
	Results []animeItem `json:"results" jsonschema:"搜索结果列表"`
}

type animeItem struct {
	Name string `json:"name" jsonschema:"动漫名称"`
	URL  string `json:"url"  jsonschema:"动漫详情页链接"`
}

func registerSearchTool(s *mcpgo.Server) {
	mcpgo.AddTool(s, &mcpgo.Tool{
		Name:        "search",
		Description: "搜索动漫，返回匹配的动漫列表（名称和详情页链接）",
	}, handleSearch)
}

func handleSearch(ctx context.Context, req *mcpgo.CallToolRequest, in searchInput) (*mcpgo.CallToolResult, searchOutput, error) {
	application := app.New()
	resultChan := application.SearchAll(in.Keyword)

	var allResults []source.Anime
	for range application.SourceManager.EnabledCount() {
		result := <-resultChan
		if result.Status == app.StatusSuccess {
			allResults = append(allResults, result.Results...)
		}
	}

	items := make([]animeItem, len(allResults))
	for i, a := range allResults {
		items[i] = animeItem{Name: a.Name, URL: a.URL}
	}

	return nil, searchOutput{Results: items}, nil
}

// --- get_episodes tool ---

type getEpisodesInput struct {
	URL string `json:"url" jsonschema:"动漫详情页URL" jsonschema_required:"true"`
}

type getEpisodesOutput struct {
	Episodes []episodeItem `json:"episodes" jsonschema:"剧集列表"`
}

type episodeItem struct {
	Name string `json:"name" jsonschema:"剧集名称"`
	URL  string `json:"url"  jsonschema:"剧集页链接"`
}

func registerGetEpisodesTool(s *mcpgo.Server) {
	mcpgo.AddTool(s, &mcpgo.Tool{
		Name:        "get_episodes",
		Description: "获取指定动漫的剧集列表",
	}, handleGetEpisodes)
}

func handleGetEpisodes(ctx context.Context, req *mcpgo.CallToolRequest, in getEpisodesInput) (*mcpgo.CallToolResult, getEpisodesOutput, error) {
	application := app.New()
	episodes, err := getEpisodesFromAnySource(application, in.URL)
	if err != nil {
		return nil, getEpisodesOutput{}, err
	}

	items := make([]episodeItem, len(episodes))
	for i, ep := range episodes {
		items[i] = episodeItem{Name: ep.Name, URL: ep.URL}
	}

	return nil, getEpisodesOutput{Episodes: items}, nil
}

// --- get_video_url tool ---

type getVideoURLInput struct {
	EpisodeURL string `json:"episode_url" jsonschema:"剧集页URL" jsonschema_required:"true"`
}

type getVideoURLOutput struct {
	VideoURL string `json:"video_url" jsonschema:"视频直链地址"`
}

func registerGetVideoURLTool(s *mcpgo.Server) {
	mcpgo.AddTool(s, &mcpgo.Tool{
		Name:        "get_video_url",
		Description: "获取指定剧集的视频直链地址",
	}, handleGetVideoURL)
}

func handleGetVideoURL(ctx context.Context, req *mcpgo.CallToolRequest, in getVideoURLInput) (*mcpgo.CallToolResult, getVideoURLOutput, error) {
	application := app.New()
	videoURL, err := getVideoURLFromAnySource(application, in.EpisodeURL)
	if err != nil {
		return nil, getVideoURLOutput{}, err
	}

	return nil, getVideoURLOutput{VideoURL: videoURL}, nil
}

// --- list_sources tool ---

type listSourcesOutput struct {
	Sources []sourceItem `json:"sources" jsonschema:"可用媒体源列表"`
}

type sourceItem struct {
	ID          string `json:"id"          jsonschema:"源唯一标识"`
	Name        string `json:"name"        jsonschema:"源名称"`
	Description string `json:"description" jsonschema:"源描述"`
	Enabled     bool   `json:"enabled"     jsonschema:"是否启用"`
}

func registerListSourcesTool(s *mcpgo.Server) {
	mcpgo.AddTool(s, &mcpgo.Tool{
		Name:        "list_sources",
		Description: "列出所有可用的媒体源及其启用状态",
	}, handleListSources)
}

func handleListSources(ctx context.Context, req *mcpgo.CallToolRequest, _ struct{}) (*mcpgo.CallToolResult, listSourcesOutput, error) {
	application := app.New()
	channels := application.SourceManager.ListChannels()

	items := make([]sourceItem, len(channels))
	for i, ch := range channels {
		items[i] = sourceItem{
			ID:          ch.ID,
			Name:        ch.Name,
			Description: ch.Description,
			Enabled:     ch.Enabled,
		}
	}

	return nil, listSourcesOutput{Sources: items}, nil
}

// --- helpers ---

// getEpisodesFromAnySource 尝试从所有启用的源获取剧集
func getEpisodesFromAnySource(application *app.App, animeURL string) ([]source.Episode, error) {
	enabled := application.SourceManager.GetEnabled()
	for _, ms := range enabled {
		ws := webselector.New(ms)
		episodes, err := ws.GetEpisodes(animeURL)
		if err == nil && len(episodes) > 0 {
			return episodes, nil
		}
	}
	return nil, fmt.Errorf("所有源均无法获取该剧集列表")
}

// getVideoURLFromAnySource 尝试从所有启用的源获取视频直链
func getVideoURLFromAnySource(application *app.App, episodeURL string) (string, error) {
	enabled := application.SourceManager.GetEnabled()
	for _, ms := range enabled {
		ws := webselector.New(ms)
		videoURL, err := ws.GetVideoURL(episodeURL)
		if err == nil && videoURL != "" {
			return videoURL, nil
		}
	}
	return "", fmt.Errorf("所有源均无法获取该视频直链")
}