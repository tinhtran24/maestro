package memory

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC) }
}

func completedEvent(id, taskID string, order int, files, tests []string) memoryevents.Event {
	return memoryevents.Event{
		V:          memoryevents.SchemaVersion,
		ID:         id,
		Type:       memoryevents.TypeTaskCompleted,
		OccurredAt: time.Date(2026, 7, 14, 8, order, 0, 0, time.UTC),
		Task: memoryevents.Task{
			ID:           taskID,
			SessionID:    "acme-" + taskID,
			ProjectID:    "acme",
			Kind:         "worker",
			Intent:       "work on " + taskID,
			TaskType:     "feature",
			Branch:       "feat/" + taskID,
			BaseSHA:      "base-" + taskID,
			HeadSHA:      "head-" + taskID,
			ChangedFiles: files,
			ChangedTests: tests,
			PRs:          []memoryevents.PRRef{{URL: "https://x/pr/" + taskID, Number: order, State: "merged"}},
		},
	}
}

type taskView struct {
	Intent, TaskType, Branch, BaseSHA, HeadSHA, EventID, PRsJSON, OccurredAt string
	Files, Tests                                                             []string
}

func snapshot(t *testing.T, store *memorydb.Store) map[string]taskView {
	t.Helper()
	ctx := context.Background()
	tasks, err := store.ListTasks(ctx, 1000)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	out := make(map[string]taskView, len(tasks))
	for _, tk := range tasks {
		files, err := store.ListFiles(ctx, tk.ID)
		if err != nil {
			t.Fatalf("ListFiles: %v", err)
		}
		tests, err := store.ListTests(ctx, tk.ID)
		if err != nil {
			t.Fatalf("ListTests: %v", err)
		}
		out[tk.ID] = taskView{
			Intent:     tk.Intent,
			TaskType:   tk.TaskType,
			Branch:     tk.Branch,
			BaseSHA:    tk.BaseSHA,
			HeadSHA:    tk.HeadSHA,
			EventID:    tk.EventID,
			PRsJSON:    tk.PRsJSON,
			OccurredAt: tk.OccurredAt.UTC().Format(time.RFC3339Nano),
			Files:      files,
			Tests:      tests,
		}
	}
	return out
}

func newProjector() *Projector {
	p := NewProjector(memoryevents.New(), nil)
	p.now = fixedClock()
	return p
}

func TestApplyThenRebuildIsIdenticalAndDoubleApplyIsNoOp(t *testing.T) {
	ctx := context.Background()
	cacheDir := t.TempDir()
	projectDir := t.TempDir()
	p := newProjector()

	store, err := memorydb.Open(cacheDir, "acme")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	evs := []memoryevents.Event{
		completedEvent("EVT-1", "TASK-1", 1, []string{"a.go", "b.go"}, []string{"a_test.go"}),
		completedEvent("EVT-2", "TASK-2", 2, []string{"b.go", "c.go"}, nil),
		completedEvent("EVT-3", "TASK-3", 3, []string{"d.go"}, []string{"d_test.go"}),
	}
	for _, ev := range evs {
		if err := p.events.Append(ctx, projectDir, ev); err != nil {
			t.Fatalf("Append %s: %v", ev.ID, err)
		}
		if err := p.Apply(ctx, store, ev); err != nil {
			t.Fatalf("Apply %s: %v", ev.ID, err)
		}
	}
	snap1 := snapshot(t, store)
	if len(snap1) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(snap1))
	}

	// Re-applying the same events must not change the projection.
	for _, ev := range evs {
		if err := p.Apply(ctx, store, ev); err != nil {
			t.Fatalf("re-Apply %s: %v", ev.ID, err)
		}
	}
	if dup := snapshot(t, store); !reflect.DeepEqual(snap1, dup) {
		t.Fatalf("double-apply changed the projection:\n%+v\n!=\n%+v", snap1, dup)
	}

	// Drop the projection entirely, then rebuild from the log alone.
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := os.RemoveAll(cacheDir); err != nil {
		t.Fatalf("wipe cache: %v", err)
	}
	rebuilt, err := memorydb.Open(cacheDir, "acme")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = rebuilt.Close() }()
	if err := p.Rebuild(ctx, rebuilt, projectDir); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if snap2 := snapshot(t, rebuilt); !reflect.DeepEqual(snap1, snap2) {
		t.Fatalf("rebuild diverged from live apply:\n%+v\n!=\n%+v", snap1, snap2)
	}

	// The offset after a full rebuild is the last event in the log.
	meta, err := rebuilt.GetMeta(ctx)
	if err != nil {
		t.Fatalf("GetMeta: %v", err)
	}
	if meta.LastEventID != "EVT-3" {
		t.Fatalf("offset after rebuild = %q, want EVT-3", meta.LastEventID)
	}
}

func TestCatchUpAppliesNewEventsExactlyOnce(t *testing.T) {
	ctx := context.Background()
	cacheDir := t.TempDir()
	projectDir := t.TempDir()
	p := newProjector()
	store, err := memorydb.Open(cacheDir, "acme")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = store.Close() }()

	for _, ev := range []memoryevents.Event{
		completedEvent("EVT-1", "TASK-1", 1, []string{"a.go"}, nil),
		completedEvent("EVT-2", "TASK-2", 2, []string{"b.go"}, nil),
	} {
		if err := p.events.Append(ctx, projectDir, ev); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	n, err := p.CatchUp(ctx, store, projectDir)
	if err != nil {
		t.Fatalf("CatchUp: %v", err)
	}
	if n != 2 {
		t.Fatalf("first CatchUp applied %d, want 2", n)
	}
	// Nothing new: a second CatchUp is a no-op.
	n, err = p.CatchUp(ctx, store, projectDir)
	if err != nil {
		t.Fatalf("CatchUp 2: %v", err)
	}
	if n != 0 {
		t.Fatalf("second CatchUp applied %d, want 0", n)
	}

	// A newly appended event is picked up on the next CatchUp.
	if err := p.events.Append(ctx, projectDir, completedEvent("EVT-3", "TASK-3", 3, nil, nil)); err != nil {
		t.Fatalf("Append: %v", err)
	}
	n, err = p.CatchUp(ctx, store, projectDir)
	if err != nil {
		t.Fatalf("CatchUp 3: %v", err)
	}
	if n != 1 {
		t.Fatalf("third CatchUp applied %d, want 1", n)
	}
	if tasks, _ := store.ListTasks(ctx, 100); len(tasks) != 3 {
		t.Fatalf("expected 3 tasks after catch-up, got %d", len(tasks))
	}
}

func TestApplyUnknownEventTypeAdvancesOffsetOnly(t *testing.T) {
	ctx := context.Background()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = store.Close() }()
	p := newProjector()

	ev := memoryevents.Event{V: 1, ID: "EVT-X", Type: "task.superseded", OccurredAt: time.Unix(1, 0).UTC()}
	if err := p.Apply(ctx, store, ev); err != nil {
		t.Fatalf("Apply unknown type: %v", err)
	}
	if tasks, _ := store.ListTasks(ctx, 10); len(tasks) != 0 {
		t.Fatalf("unknown event should not create a task, got %d", len(tasks))
	}
	meta, _ := store.GetMeta(ctx)
	if meta.LastEventID != "EVT-X" {
		t.Fatalf("offset should advance past an unknown event, got %q", meta.LastEventID)
	}
}

func TestApplyRejectsMalformedEvents(t *testing.T) {
	ctx := context.Background()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = store.Close() }()
	p := newProjector()

	if err := p.Apply(ctx, store, memoryevents.Event{Type: memoryevents.TypeTaskCompleted}); err == nil {
		t.Error("expected error for an event with no id")
	}
	// task.completed with an empty task id cannot be projected.
	ev := memoryevents.Event{ID: "EVT-1", Type: memoryevents.TypeTaskCompleted, OccurredAt: time.Unix(1, 0).UTC()}
	if err := p.Apply(ctx, store, ev); err == nil {
		t.Error("expected error for a task.completed event with no task id")
	}
}
