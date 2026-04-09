package app

import (
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
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
	SourceName string
	Status     SearchStatus
	Duration   time.Duration
	Results    []source.Anime
	Error      error
}
