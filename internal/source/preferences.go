package source

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// DoctorStatus 表示渠道诊断结果。
type DoctorStatus string

const (
	DoctorStatusUnknown      DoctorStatus = "unknown"
	DoctorStatusSearchError  DoctorStatus = "search_error"
	DoctorStatusNoSearchHits DoctorStatus = "no_search_hits"
	DoctorStatusEpisodeError DoctorStatus = "episode_error"
	DoctorStatusNoEpisodes   DoctorStatus = "no_episodes"
	DoctorStatusVideoError   DoctorStatus = "video_error"
	DoctorStatusVideoReady   DoctorStatus = "video_ready"
)

// SourceDoctorCase 表示一次单样本动漫诊断结果。
type SourceDoctorCase struct {
	CheckedAt    string       `json:"checkedAt,omitempty"`
	Keyword      string       `json:"keyword,omitempty"`
	Status       DoctorStatus `json:"status,omitempty"`
	SearchHits   int          `json:"searchHits,omitempty"`
	EpisodeCount int          `json:"episodeCount,omitempty"`
	VideoReady   bool         `json:"videoReady,omitempty"`
	DurationMS   int64        `json:"durationMs,omitempty"`
	Message      string       `json:"message,omitempty"`
}

// SourceDoctorSnapshot 表示一次完整 doctor 诊断结果。
type SourceDoctorSnapshot struct {
	CheckedAt      string             `json:"checkedAt,omitempty"`
	SuccessfulRuns int                `json:"successfulRuns,omitempty"`
	TotalRuns      int                `json:"totalRuns,omitempty"`
	AverageMS      int64              `json:"averageMs,omitempty"`
	Priority       int                `json:"priority,omitempty"`
	AutoDisabled   bool               `json:"autoDisabled,omitempty"`
	Summary        string             `json:"summary,omitempty"`
	Cases          []SourceDoctorCase `json:"cases,omitempty"`
}

// SourcePreference 表示单个渠道的本地偏好。
type SourcePreference struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Enabled    bool                  `json:"enabled"`
	Priority   int                   `json:"priority,omitempty"`
	LastDoctor *SourceDoctorSnapshot `json:"lastDoctor,omitempty"`
}

// SourcePreferenceFile 表示渠道偏好文件。
type SourcePreferenceFile struct {
	Sources []SourcePreference `json:"sources"`
}

// SourceChannelState 表示加载后的渠道状态，用于 CLI 和 TUI 展示。
type SourceChannelState struct {
	ID          string
	Name        string
	Description string
	Enabled     bool
	Priority    int
	LastDoctor  *SourceDoctorSnapshot
}

// ChannelDoctorUpdate 表示要写回的渠道诊断结果。
type ChannelDoctorUpdate struct {
	ID       string
	Name     string
	Priority int
	Enabled  bool
	Snapshot SourceDoctorSnapshot
}

type sourcePreferenceStore struct {
	path string
}

// PreferenceID 返回用于渠道偏好持久化的稳定 ID。
func (ms MediaSource) PreferenceID() string {
	return sourcePreferenceID(ms)
}

func newSourcePreferenceStore(path string) *sourcePreferenceStore {
	return &sourcePreferenceStore{path: path}
}

func (s *sourcePreferenceStore) Load() (*SourcePreferenceFile, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	var file SourcePreferenceFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *sourcePreferenceStore) Save(file SourcePreferenceFile) error {
	configDir := filepath.Dir(s.path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func sourcePreferenceID(ms MediaSource) string {
	factoryID := strings.TrimSpace(ms.FactoryID)
	name := strings.TrimSpace(ms.Arguments.Name)

	switch {
	case factoryID == "" && name == "":
		return "unknown"
	case factoryID == "":
		return "name:" + name
	case name == "":
		return "factory:" + factoryID
	default:
		return "factory:" + factoryID + "|name:" + name
	}
}

func syncSourcePreferences(sources []MediaSource, existing SourcePreferenceFile) SourcePreferenceFile {
	existingByID := make(map[string]SourcePreference, len(existing.Sources))
	for _, pref := range existing.Sources {
		existingByID[pref.ID] = cloneSourcePreference(pref)
	}

	next := SourcePreferenceFile{
		Sources: make([]SourcePreference, 0, len(sources)),
	}
	for _, ms := range sources {
		id := sourcePreferenceID(ms)
		pref, ok := existingByID[id]
		if !ok {
			pref = SourcePreference{
				ID:      id,
				Name:    displaySourcePreferenceName(ms),
				Enabled: true,
			}
		}
		pref.ID = id
		pref.Name = displaySourcePreferenceName(ms)
		next.Sources = append(next.Sources, cloneSourcePreference(pref))
	}

	return next
}

func displaySourcePreferenceName(ms MediaSource) string {
	name := strings.TrimSpace(ms.Arguments.Name)
	if name != "" {
		return name
	}
	return strings.TrimSpace(ms.FactoryID)
}

func cloneSourcePreference(pref SourcePreference) SourcePreference {
	cloned := pref
	cloned.LastDoctor = cloneDoctorSnapshot(pref.LastDoctor)
	return cloned
}

func cloneDoctorSnapshot(snapshot *SourceDoctorSnapshot) *SourceDoctorSnapshot {
	if snapshot == nil {
		return nil
	}
	cloned := *snapshot
	if len(snapshot.Cases) > 0 {
		cloned.Cases = make([]SourceDoctorCase, len(snapshot.Cases))
		copy(cloned.Cases, snapshot.Cases)
	}
	return &cloned
}
