package app

import (
	"context"
	"errors"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

// SearchStatus 搜索状态
type SearchStatus int

const (
	StatusSearching SearchStatus = iota
	StatusSuccess
	StatusTimeout
	StatusError
)

// SourceSearchResult 单个源的搜索结果
type SourceSearchResult struct {
	SourceName     string
	SourcePriority int
	Status         SearchStatus
	Duration       time.Duration
	Results        []source.Anime
	Error          error
}

// SearchAll 并发搜索所有源
func (a *App) SearchAll(keyword string) <-chan SourceSearchResult {
	sources := a.SourceManager.GetEnabled()
	resultChan := make(chan SourceSearchResult, len(sources))

	for i := range sources {
		go func(ms source.MediaSource) {
			start := time.Now()
			ws := webselector.New(ms)

			results, err := ws.Search(keyword)
			duration := time.Since(start)

			result := SourceSearchResult{
				SourceName:     ms.Arguments.Name,
				SourcePriority: a.SourceManager.GetChannelPriorityByName(ms.Arguments.Name),
				Duration:       duration,
			}

			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					result.Status = StatusTimeout
				} else {
					result.Status = StatusError
					result.Error = err
				}
			} else {
				result.Status = StatusSuccess
				result.Results = results
			}

			resultChan <- result
		}(sources[i])
	}

	return resultChan
}
