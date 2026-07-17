package memory

import (
	"context"
	"testing"
	"time"

	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

func seedTask(t *testing.T, store *memorydb.Store, id string, occ time.Time, files, tests []string) {
	t.Helper()
	ctx := context.Background()
	if err := store.UpsertTask(ctx, memorydb.Task{
		ID: id, EventID: "e-" + id, SessionID: "s-" + id, ProjectID: "acme", OccurredAt: occ,
	}); err != nil {
		t.Fatalf("seed task %s: %v", id, err)
	}
	if err := store.ReplaceFiles(ctx, id, files); err != nil {
		t.Fatalf("seed files %s: %v", id, err)
	}
	if err := store.ReplaceTests(ctx, id, tests); err != nil {
		t.Fatalf("seed tests %s: %v", id, err)
	}
}

// openResolverStore seeds three historical tasks: A and B both touch the shared
// file, B additionally touches a second queried file (so it ranks above A), and
// C is unrelated.
func openResolverStore(t *testing.T) *memorydb.Store {
	t.Helper()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	base := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	seedTask(t, store, "TASK-A", base, []string{"shared.go", "a.go"}, nil)
	seedTask(t, store, "TASK-B", base.Add(time.Hour), []string{"shared.go", "x.go", "b.go"}, nil)
	seedTask(t, store, "TASK-C", base.Add(2*time.Hour), []string{"unrelated.go"}, nil)
	return store
}

func ids(rts []RelatedTask) []string {
	out := make([]string, len(rts))
	for i, rt := range rts {
		out[i] = rt.TaskID
	}
	return out
}

func TestResolverReturnsRankedSharersNotUnrelated(t *testing.T) {
	store := openResolverStore(t)
	r := NewResolver()

	got, err := r.Related(context.Background(), store, Query{Files: []string{"shared.go", "x.go"}})
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	// B overlaps two queried files, A overlaps one; C overlaps none and is dropped.
	if want := []string{"TASK-B", "TASK-A"}; !equalStrings(ids(got), want) {
		t.Fatalf("related = %v, want %v", ids(got), want)
	}
	if got[0].Score != 2 || got[1].Score != 1 {
		t.Fatalf("scores = %d,%d want 2,1", got[0].Score, got[1].Score)
	}
	if got[0].SharedFiles != 2 {
		t.Fatalf("TASK-B shared files = %d, want 2", got[0].SharedFiles)
	}
}

func TestResolverWeightsTestOverlap(t *testing.T) {
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()
	base := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	seedTask(t, store, "TASK-A", base, []string{"shared.go"}, nil)
	seedTask(t, store, "TASK-B", base.Add(time.Hour), []string{"shared.go"}, []string{"shared_test.go"})

	got, err := NewResolver().Related(context.Background(), store, Query{
		Files: []string{"shared.go"}, Tests: []string{"shared_test.go"},
	})
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	// B shares a file and a test (score 2); A shares only the file (score 1).
	if want := []string{"TASK-B", "TASK-A"}; !equalStrings(ids(got), want) {
		t.Fatalf("related = %v, want %v", ids(got), want)
	}
	if got[0].SharedTests != 1 {
		t.Fatalf("TASK-B shared tests = %d, want 1", got[0].SharedTests)
	}
}

func TestResolverCapsToLimit(t *testing.T) {
	store := openResolverStore(t)
	got, err := NewResolver().Related(context.Background(), store, Query{Files: []string{"shared.go", "x.go"}, Limit: 1})
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	if len(got) != 1 || got[0].TaskID != "TASK-B" {
		t.Fatalf("limit=1 should return only the top task, got %v", ids(got))
	}
}

func TestResolverExcludesSelf(t *testing.T) {
	store := openResolverStore(t)
	got, err := NewResolver().Related(context.Background(), store, Query{
		Files: []string{"shared.go", "x.go"}, ExcludeTaskID: "TASK-B",
	})
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	if want := []string{"TASK-A"}; !equalStrings(ids(got), want) {
		t.Fatalf("excluding TASK-B should leave %v, got %v", want, ids(got))
	}
}

func TestResolverDedupesQueryPaths(t *testing.T) {
	store := openResolverStore(t)
	// A duplicated query path must not inflate the shared count.
	got, err := NewResolver().Related(context.Background(), store, Query{Files: []string{"shared.go", "shared.go"}})
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	for _, rt := range got {
		if rt.SharedFiles != 1 {
			t.Fatalf("task %s shared files = %d, want 1 (dedup)", rt.TaskID, rt.SharedFiles)
		}
	}
}

func TestResolverEmptyQueryReturnsNothing(t *testing.T) {
	store := openResolverStore(t)
	got, err := NewResolver().Related(context.Background(), store, Query{})
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("empty query should return nothing, got %v", ids(got))
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
