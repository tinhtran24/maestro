package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

type fakeProjects struct {
	rec domain.ProjectRecord
	ok  bool
	err error
}

func (f fakeProjects) GetProject(_ context.Context, _ string) (domain.ProjectRecord, bool, error) {
	return f.rec, f.ok, f.err
}

func newRecorder(t *testing.T, projectDir, cacheDir string, projects ProjectLookup, diff Differ) *Recorder {
	t.Helper()
	events := memoryevents.New()
	builder := NewBuilder(diff, nil)
	projector := NewProjector(events, nil)
	return NewRecorder(builder, events, projector, projects, nil, cacheDir, nil)
}

func countEvents(t *testing.T, projectDir string) []memoryevents.Event {
	t.Helper()
	var got []memoryevents.Event
	if err := memoryevents.New().Iterate(context.Background(), projectDir, "", func(ev memoryevents.Event) error {
		got = append(got, ev)
		return nil
	}); err != nil {
		t.Fatalf("Iterate: %v", err)
	}
	return got
}

func TestRecordCompletionCapturesWorkerAndOrchestrator(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	cacheDir := t.TempDir()
	projects := fakeProjects{rec: domain.ProjectRecord{ID: "acme", Path: projectDir}, ok: true}
	diff := &fakeDiffer{files: []string{"a.go", "a_test.go"}}
	r := newRecorder(t, projectDir, cacheDir, projects, diff)

	worker := domain.SessionRecord{ID: "acme-1", ProjectID: "acme", Kind: domain.KindWorker, Metadata: domain.SessionMetadata{Branch: "feat/a", Prompt: "worker task"}}
	orch := domain.SessionRecord{ID: "acme-2", ProjectID: "acme", Kind: domain.KindOrchestrator, Metadata: domain.SessionMetadata{Branch: "feat/b", Prompt: "orchestrator task"}}
	r.RecordCompletion(ctx, worker)
	r.RecordCompletion(ctx, orch)

	evs := countEvents(t, projectDir)
	if len(evs) != 2 {
		t.Fatalf("expected 2 events (worker + orchestrator), got %d", len(evs))
	}
	kinds := map[string]bool{}
	for _, ev := range evs {
		kinds[ev.Task.Kind] = true
	}
	if !kinds["worker"] || !kinds["orchestrator"] {
		t.Fatalf("expected both worker and orchestrator events, got %v", kinds)
	}

	// The projection was folded and split the test out from the source file.
	store, err := memorydb.Open(cacheDir, "acme")
	if err != nil {
		t.Fatalf("open projection: %v", err)
	}
	defer func() { _ = store.Close() }()
	tasks, err := store.ListTasks(ctx, 100)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 projected tasks, got %d", len(tasks))
	}
	files, _ := store.ListFiles(ctx, "acme-1")
	testsFiles, _ := store.ListTests(ctx, "acme-1")
	if len(files) != 1 || files[0] != "a.go" {
		t.Errorf("projected files = %v, want [a.go]", files)
	}
	if len(testsFiles) != 1 || testsFiles[0] != "a_test.go" {
		t.Errorf("projected tests = %v, want [a_test.go]", testsFiles)
	}
}

func TestRecordCompletionIsBestEffortOnProjectLookupFailure(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	cacheDir := t.TempDir()
	projects := fakeProjects{err: errors.New("db down")}
	r := newRecorder(t, projectDir, cacheDir, projects, &fakeDiffer{})

	// Must not panic and must write nothing.
	r.RecordCompletion(ctx, domain.SessionRecord{ID: "acme-1", ProjectID: "acme", Metadata: domain.SessionMetadata{Branch: "feat/a"}})
	if evs := countEvents(t, projectDir); len(evs) != 0 {
		t.Fatalf("expected no events on lookup failure, got %d", len(evs))
	}
}

func TestRecordCompletionSkipsUnknownProject(t *testing.T) {
	ctx := context.Background()
	projectDir := t.TempDir()
	cacheDir := t.TempDir()
	projects := fakeProjects{ok: false}
	r := newRecorder(t, projectDir, cacheDir, projects, &fakeDiffer{})
	r.RecordCompletion(ctx, domain.SessionRecord{ID: "acme-1", ProjectID: "acme"})
	if evs := countEvents(t, projectDir); len(evs) != 0 {
		t.Fatalf("expected no events for an unknown project, got %d", len(evs))
	}
}
