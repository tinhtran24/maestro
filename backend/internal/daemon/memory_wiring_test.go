package daemon

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/tinhtran24/maestro/backend/internal/config"
	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
	"github.com/tinhtran24/maestro/backend/internal/storage/sqlite"
)

// TestBuildMemoryRecorder_CapturesCompletion exercises the real wiring: the
// recorder resolves the project checkout from the store, appends a completion
// event to that project's events.jsonl, and folds it into the projection under
// <DataDir>/memory/<projectID>. It uses a branchless session so the git differ
// is skipped and the test needs no real repository.
func TestBuildMemoryRecorder_CapturesCompletion(t *testing.T) {
	ctx := context.Background()
	store, err := sqlite.Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	projectDir := t.TempDir()
	if err := store.UpsertProject(ctx, domain.ProjectRecord{ID: "mer", Path: projectDir}); err != nil {
		t.Fatalf("register project: %v", err)
	}

	cfg := config.Config{DataDir: t.TempDir()}
	rec := buildMemoryRecorder(cfg, store, slog.Default())
	if rec == nil {
		t.Fatal("buildMemoryRecorder returned nil")
	}

	rec.RecordCompletion(ctx, domain.SessionRecord{
		ID:        "mer-1",
		ProjectID: "mer",
		Kind:      domain.KindWorker,
		Metadata:  domain.SessionMetadata{Prompt: "do the thing"},
	})

	// The committed event log lives inside the project checkout.
	logPath := filepath.Join(projectDir, ".maestro", "memory", "events.jsonl")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected events.jsonl at %s: %v", logPath, err)
	}

	// The projection lives under <DataDir>/memory and holds the folded task.
	proj, err := memorydb.Open(filepath.Join(cfg.DataDir, "memory"), "mer")
	if err != nil {
		t.Fatalf("open projection: %v", err)
	}
	defer func() { _ = proj.Close() }()
	tasks, err := proj.ListTasks(ctx, 10)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "mer-1" || tasks[0].Intent != "do the thing" {
		t.Fatalf("projection did not capture the completion: %+v", tasks)
	}
}
