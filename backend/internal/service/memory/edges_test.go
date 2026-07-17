package memory

import (
	"context"
	"testing"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

// TestApplyPopulatesSharedPathEdges verifies the task graph is materialised as
// events are folded: the second task, which shares a file with the first, gets
// an outgoing shared_path edge to it, and a rebuild reproduces the same graph.
func TestApplyPopulatesSharedPathEdges(t *testing.T) {
	ctx := context.Background()
	cacheDir := t.TempDir()
	projectDir := t.TempDir()
	p := newProjector()
	store, err := memorydb.Open(cacheDir, "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	evs := []memoryevents.Event{
		completedEvent("EVT-1", "TASK-1", 1, []string{"a.go", "shared.go"}, nil),
		completedEvent("EVT-2", "TASK-2", 2, []string{"shared.go", "c.go"}, nil),
	}
	for _, ev := range evs {
		if err := p.events.Append(ctx, projectDir, ev); err != nil {
			t.Fatalf("append: %v", err)
		}
		if err := p.Apply(ctx, store, ev); err != nil {
			t.Fatalf("apply %s: %v", ev.ID, err)
		}
	}

	// TASK-2 was applied second, so it discovered the overlap and owns the edge.
	edges, err := store.ListEdgesFrom(ctx, "TASK-2")
	if err != nil {
		t.Fatalf("list edges: %v", err)
	}
	if len(edges) != 1 || edges[0].DstTaskID != "TASK-1" || edges[0].Relation != relationShared || edges[0].Confidence != 1 {
		t.Fatalf("TASK-2 edges = %+v, want one shared_path->TASK-1 confidence 1", edges)
	}
	// TASK-1 had no prior task to link to when it was applied.
	if first, _ := store.ListEdgesFrom(ctx, "TASK-1"); len(first) != 0 {
		t.Fatalf("TASK-1 should have no outgoing edges, got %+v", first)
	}

	// A rebuild reproduces the same edge.
	if err := p.Rebuild(ctx, store, projectDir); err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	edges, _ = store.ListEdgesFrom(ctx, "TASK-2")
	if len(edges) != 1 || edges[0].DstTaskID != "TASK-1" {
		t.Fatalf("after rebuild TASK-2 edges = %+v, want one ->TASK-1", edges)
	}
	_ = store.Close()
}

// TestApplyRefreshesEdgesOnReapply checks a re-applied task's edges track its
// current file set rather than accumulating stale links.
func TestApplyRefreshesEdgesOnReapply(t *testing.T) {
	ctx := context.Background()
	p := newProjector()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()

	first := completedEvent("EVT-1", "TASK-1", 1, []string{"shared.go"}, nil)
	second := completedEvent("EVT-2", "TASK-2", 2, []string{"shared.go"}, nil)
	for _, ev := range []memoryevents.Event{first, second} {
		if err := p.Apply(ctx, store, ev); err != nil {
			t.Fatalf("apply: %v", err)
		}
	}
	if edges, _ := store.ListEdgesFrom(ctx, "TASK-2"); len(edges) != 1 {
		t.Fatalf("expected one edge before re-apply, got %+v", edges)
	}

	// Re-apply TASK-2 with a disjoint file set: its edge to TASK-1 must drop.
	moved := completedEvent("EVT-2b", "TASK-2", 2, []string{"elsewhere.go"}, nil)
	if err := p.Apply(ctx, store, moved); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	if edges, _ := store.ListEdgesFrom(ctx, "TASK-2"); len(edges) != 0 {
		t.Fatalf("re-applied TASK-2 no longer shares paths; edges should be empty, got %+v", edges)
	}
}
