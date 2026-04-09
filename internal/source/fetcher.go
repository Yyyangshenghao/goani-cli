package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type sourceFetcher struct {
	client *http.Client
}

func newSourceFetcher(timeout time.Duration) *sourceFetcher {
	return &sourceFetcher{
		client: &http.Client{Timeout: timeout},
	}
}

func (f *sourceFetcher) FetchURL(url string) ([]MediaSource, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg.ExportedMediaSourceDataList.MediaSources, nil
}

func (f *sourceFetcher) FetchSubscriptions(subscriptions []Subscription) ([]Subscription, []MediaSource, bool) {
	if len(subscriptions) == 0 {
		return []Subscription{}, []MediaSource{}, true
	}

	refreshed := cloneSubscriptions(subscriptions)
	sources := make([]MediaSource, 0)
	successCount := 0

	for i := range refreshed {
		fetchedSources, err := f.FetchURL(refreshed[i].URL)
		if err != nil {
			continue
		}
		refreshed[i].UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
		sources = append(sources, fetchedSources...)
		successCount++
	}

	return refreshed, sources, successCount > 0
}
