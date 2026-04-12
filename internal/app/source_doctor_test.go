package app

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Yyyangshenghao/goani-cli/internal/source"
	"github.com/Yyyangshenghao/goani-cli/internal/source/webselector"
)

func TestRunDoctorForSourceRunsCasesConcurrently(t *testing.T) {
	originalKeywords := defaultDoctorKeywords
	originalRunner := doctorCaseRunner
	t.Cleanup(func() {
		defaultDoctorKeywords = originalKeywords
		doctorCaseRunner = originalRunner
	})

	defaultDoctorKeywords = []string{"a", "b", "c", "d", "e"}

	var currentConcurrent int32
	var maxConcurrent int32
	var mu sync.Mutex
	startedKeywords := make([]string, 0, len(defaultDoctorKeywords))

	doctorCaseRunner = func(_ *webselector.WebSelectorSource, keyword string) source.SourceDoctorCase {
		mu.Lock()
		startedKeywords = append(startedKeywords, keyword)
		mu.Unlock()

		now := atomic.AddInt32(&currentConcurrent, 1)
		for {
			observed := atomic.LoadInt32(&maxConcurrent)
			if now <= observed || atomic.CompareAndSwapInt32(&maxConcurrent, observed, now) {
				break
			}
		}

		time.Sleep(80 * time.Millisecond)
		atomic.AddInt32(&currentConcurrent, -1)

		return source.SourceDoctorCase{
			Keyword:    keyword,
			Status:     source.DoctorStatusVideoReady,
			VideoReady: true,
			DurationMS: 80,
		}
	}

	start := time.Now()
	result := runDoctorForSource(source.MediaSource{
		FactoryID: "factory-a",
		Arguments: source.Arguments{
			Name: "测试源",
		},
	})
	elapsed := time.Since(start)

	if len(startedKeywords) != len(defaultDoctorKeywords) {
		t.Fatalf("unexpected case count: got %d want %d", len(startedKeywords), len(defaultDoctorKeywords))
	}
	if atomic.LoadInt32(&maxConcurrent) <= 1 {
		t.Fatalf("expected doctor cases to run concurrently, max concurrent = %d", maxConcurrent)
	}

	serialCost := 80 * time.Millisecond * time.Duration(len(defaultDoctorKeywords))
	if elapsed >= serialCost-50*time.Millisecond {
		t.Fatalf("expected concurrent execution to beat near-serial runtime: elapsed=%v serial=%v", elapsed, serialCost)
	}

	if result.Snapshot.SuccessfulRuns != len(defaultDoctorKeywords) {
		t.Fatalf("unexpected successful runs: got %d want %d", result.Snapshot.SuccessfulRuns, len(defaultDoctorKeywords))
	}
	if !result.Enabled {
		t.Fatalf("expected source to remain enabled after all cases succeed")
	}
}
