package source

import "testing"

func TestSyncSourcePreferencesPreservesExistingState(t *testing.T) {
	first := MediaSource{
		FactoryID: "factory-a",
		Arguments: Arguments{
			Name:        "源A",
			Description: "测试源 A",
		},
	}
	second := MediaSource{
		FactoryID: "factory-b",
		Arguments: Arguments{
			Name:        "源B",
			Description: "测试源 B",
		},
	}

	existing := SourcePreferenceFile{
		Sources: []SourcePreference{
			{
				ID:       sourcePreferenceID(first),
				Name:     "源A",
				Enabled:  false,
				Priority: 420,
				LastDoctor: &SourceDoctorSnapshot{
					CheckedAt:      "2026-04-12 10:00:00",
					SuccessfulRuns: 5,
					TotalRuns:      5,
					Priority:       420,
					Cases: []SourceDoctorCase{
						{
							Status:     DoctorStatusVideoReady,
							CheckedAt:  "2026-04-12 10:00:00",
							Keyword:    "葬送的芙莉莲",
							SearchHits: 3,
							VideoReady: true,
						},
					},
				},
			},
			{
				ID:      "stale",
				Name:    "过期源",
				Enabled: false,
			},
		},
	}

	next := syncSourcePreferences([]MediaSource{first, second}, existing)
	if len(next.Sources) != 2 {
		t.Fatalf("unexpected source preference count: got %d want %d", len(next.Sources), 2)
	}

	if next.Sources[0].ID != sourcePreferenceID(first) {
		t.Fatalf("unexpected first source id: got %s want %s", next.Sources[0].ID, sourcePreferenceID(first))
	}
	if next.Sources[0].Enabled {
		t.Fatalf("expected first source to remain disabled")
	}
	if next.Sources[0].Priority != 420 {
		t.Fatalf("expected first source to preserve priority")
	}
	if next.Sources[0].LastDoctor == nil || next.Sources[0].LastDoctor.SuccessfulRuns != 5 {
		t.Fatalf("expected first source to preserve last doctor status")
	}

	if next.Sources[1].ID != sourcePreferenceID(second) {
		t.Fatalf("unexpected second source id: got %s want %s", next.Sources[1].ID, sourcePreferenceID(second))
	}
	if !next.Sources[1].Enabled {
		t.Fatalf("expected new source to default to enabled")
	}
	if next.Sources[1].LastDoctor != nil {
		t.Fatalf("expected new source to start without doctor snapshot")
	}
}
