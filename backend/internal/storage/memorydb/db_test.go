package memorydb

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTemp(t *testing.T, projectID string) (*Store, string) {
	t.Helper()
	cacheDir := t.TempDir()
	s, err := Open(cacheDir, projectID)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, cacheDir
}

func TestOpenPlacesDBUnderCacheDir(t *testing.T) {
	s, cacheDir := openTemp(t, "acme")
	want := filepath.Join(cacheDir, "acme", "memory.db")
	if s.Path() != want {
		t.Fatalf("db path = %q, want %q", s.Path(), want)
	}
	// The projection must live under the provided cache root and nowhere near the
	// app database file name.
	if !strings.HasPrefix(s.Path(), cacheDir) {
		t.Fatalf("db path %q escaped cache dir %q", s.Path(), cacheDir)
	}
	if strings.Contains(s.Path(), "maestro.db") {
		t.Fatalf("projection must not collide with the app db: %q", s.Path())
	}
}

func TestUpsertTaskWithFileEdgesRoundTrip(t *testing.T) {
	s, _ := openTemp(t, "acme")
	ctx := context.Background()

	task := Task{
		ID:         "TASK-001",
		EventID:    "EVT-1",
		SessionID:  "acme-7",
		ProjectID:  "acme",
		Kind:       "worker",
		Intent:     "add the widget",
		TaskType:   "feature",
		Branch:     "feat/widget",
		BaseSHA:    "aaaa",
		HeadSHA:    "bbbb",
		OccurredAt: time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC),
	}
	if err := s.UpsertTask(ctx, task); err != nil {
		t.Fatalf("UpsertTask: %v", err)
	}
	if err := s.ReplaceFiles(ctx, task.ID, []string{"internal/widget.go", "internal/api.go"}); err != nil {
		t.Fatalf("ReplaceFiles: %v", err)
	}
	if err := s.ReplaceTests(ctx, task.ID, []string{"internal/widget_test.go"}); err != nil {
		t.Fatalf("ReplaceTests: %v", err)
	}

	got, ok, err := s.GetTask(ctx, task.ID)
	if err != nil || !ok {
		t.Fatalf("GetTask: ok=%v err=%v", ok, err)
	}
	if got.Intent != "add the widget" || got.TaskType != "feature" || got.BaseSHA != "aaaa" {
		t.Fatalf("task fields not preserved: %+v", got)
	}
	// Empty PR/decision blobs default to JSON empty arrays, not empty strings.
	if got.PRsJSON != "[]" || got.DecisionsJSON != "[]" {
		t.Fatalf("expected empty json arrays, got prs=%q decisions=%q", got.PRsJSON, got.DecisionsJSON)
	}

	files, err := s.ListFiles(ctx, task.ID)
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) != 2 || files[0] != "internal/api.go" || files[1] != "internal/widget.go" {
		t.Fatalf("files = %v, want sorted [api widget]", files)
	}
	tests, err := s.ListTests(ctx, task.ID)
	if err != nil {
		t.Fatalf("ListTests: %v", err)
	}
	if len(tests) != 1 || tests[0] != "internal/widget_test.go" {
		t.Fatalf("tests = %v", tests)
	}
}

func TestReplaceFilesIsIdempotent(t *testing.T) {
	s, _ := openTemp(t, "acme")
	ctx := context.Background()
	mustTask(t, s, "TASK-1")

	for i := 0; i < 3; i++ {
		if err := s.ReplaceFiles(ctx, "TASK-1", []string{"a.go", "b.go"}); err != nil {
			t.Fatalf("ReplaceFiles iter %d: %v", i, err)
		}
	}
	files, err := s.ListFiles(ctx, "TASK-1")
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files after repeated replace, got %v", files)
	}
	// A subsequent replace with a smaller set discards the old edges.
	if err := s.ReplaceFiles(ctx, "TASK-1", []string{"a.go"}); err != nil {
		t.Fatalf("ReplaceFiles shrink: %v", err)
	}
	files, _ = s.ListFiles(ctx, "TASK-1")
	if len(files) != 1 || files[0] != "a.go" {
		t.Fatalf("expected [a.go] after shrink, got %v", files)
	}
}

func TestUpsertTaskReplacesExisting(t *testing.T) {
	s, _ := openTemp(t, "acme")
	ctx := context.Background()
	base := Task{ID: "TASK-1", EventID: "EVT-1", SessionID: "acme-1", ProjectID: "acme", Intent: "v1", OccurredAt: time.Unix(1, 0).UTC()}
	if err := s.UpsertTask(ctx, base); err != nil {
		t.Fatalf("UpsertTask: %v", err)
	}
	base.Intent = "v2"
	base.EventID = "EVT-2"
	if err := s.UpsertTask(ctx, base); err != nil {
		t.Fatalf("UpsertTask replace: %v", err)
	}
	got, _, _ := s.GetTask(ctx, "TASK-1")
	if got.Intent != "v2" || got.EventID != "EVT-2" {
		t.Fatalf("upsert did not replace: %+v", got)
	}
}

func TestMetaRoundTrip(t *testing.T) {
	s, _ := openTemp(t, "acme")
	ctx := context.Background()

	m, err := s.GetMeta(ctx)
	if err != nil {
		t.Fatalf("GetMeta empty: %v", err)
	}
	if m.LastEventID != "" {
		t.Fatalf("empty projection should have blank meta, got %q", m.LastEventID)
	}
	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	if err := s.SetMeta(ctx, "EVT-9", now); err != nil {
		t.Fatalf("SetMeta: %v", err)
	}
	m, _ = s.GetMeta(ctx)
	if m.LastEventID != "EVT-9" {
		t.Fatalf("meta = %+v, want EVT-9", m)
	}
}

func TestGetTaskMissing(t *testing.T) {
	s, _ := openTemp(t, "acme")
	_, ok, err := s.GetTask(context.Background(), "nope")
	if err != nil {
		t.Fatalf("GetTask missing: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for missing task")
	}
}

func TestOpenRejectsUnsafeProjectID(t *testing.T) {
	for _, id := range []string{"", "..", "a/b"} {
		if _, err := Open(t.TempDir(), id); err == nil {
			t.Errorf("Open should reject project id %q", id)
		}
	}
}

func mustTask(t *testing.T, s *Store, id string) {
	t.Helper()
	if err := s.UpsertTask(context.Background(), Task{
		ID: id, EventID: "EVT-" + id, SessionID: "acme-1", ProjectID: "acme", OccurredAt: time.Unix(1, 0).UTC(),
	}); err != nil {
		t.Fatalf("seed task %s: %v", id, err)
	}
}
