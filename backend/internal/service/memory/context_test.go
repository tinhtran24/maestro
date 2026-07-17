package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

func seedRich(t *testing.T, store *memorydb.Store, id, taskType string, occ time.Time, files, tests []string, prsJSON, decisionsJSON string) {
	t.Helper()
	ctx := context.Background()
	if err := store.UpsertTask(ctx, memorydb.Task{
		ID: id, EventID: "e-" + id, SessionID: "s-" + id, ProjectID: "acme",
		TaskType: taskType, Intent: "work on " + id, Branch: taskType + "/" + id,
		OccurredAt: occ, PRsJSON: prsJSON, DecisionsJSON: decisionsJSON,
	}); err != nil {
		t.Fatalf("seed %s: %v", id, err)
	}
	if err := store.ReplaceFiles(ctx, id, files); err != nil {
		t.Fatalf("files %s: %v", id, err)
	}
	if err := store.ReplaceTests(ctx, id, tests); err != nil {
		t.Fatalf("tests %s: %v", id, err)
	}
}

func TestContextPackStaysUnderBudgetAndReportsDrops(t *testing.T) {
	ctx := context.Background()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()

	base := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	const n = 60
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("T-%02d", i)
		seedRich(t, store, id, "feature", base.Add(time.Duration(i)*time.Minute),
			[]string{"shared.go", fmt.Sprintf("file-%d.go", i)}, nil, "[]", "[]")
	}

	budget := 200
	cr := NewContextResolver(budget, nil)
	pack, err := cr.Pack(ctx, store, Query{Files: []string{"shared.go"}}, RoleCoder)
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if pack.EstimatedTokens > budget {
		t.Fatalf("pack estimate %d exceeds budget %d", pack.EstimatedTokens, budget)
	}
	if len(pack.RelatedTasks) == 0 || len(pack.RelatedTasks) >= n {
		t.Fatalf("expected the pack to be trimmed, included %d of %d", len(pack.RelatedTasks), n)
	}
	if len(pack.Dropped) == 0 {
		t.Fatal("expected dropped items to be reported, not silently omitted")
	}
	if len(pack.RelatedTasks)+len(pack.Dropped) != n {
		t.Fatalf("included(%d) + dropped(%d) should account for all %d candidates", len(pack.RelatedTasks), len(pack.Dropped), n)
	}
	for _, d := range pack.Dropped {
		if d.Reason != "token_budget" || d.Kind != "related_task" {
			t.Fatalf("unexpected dropped item: %+v", d)
		}
	}
}

func TestContextPackTinyBudgetDropsEverythingButStaysUnderCap(t *testing.T) {
	ctx := context.Background()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()
	seedRich(t, store, "T1", "feature", time.Unix(1, 0).UTC(), []string{"shared.go"}, nil, "[]", "[]")

	pack, err := NewContextResolver(1, nil).Pack(ctx, store, Query{Files: []string{"shared.go"}}, RoleCoder)
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if pack.EstimatedTokens > 1 {
		t.Fatalf("estimate %d exceeds tiny budget", pack.EstimatedTokens)
	}
	if len(pack.RelatedTasks) != 0 || len(pack.Dropped) != 1 {
		t.Fatalf("tiny budget should drop the only task: included=%d dropped=%d", len(pack.RelatedTasks), len(pack.Dropped))
	}
}

func TestContextPackRoleSlices(t *testing.T) {
	ctx := context.Background()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()
	seedRich(t, store, "T1", "feature", time.Unix(10, 0).UTC(),
		[]string{"shared.go", "impl.go"}, []string{"shared_test.go"},
		`[{"url":"https://pr/1","number":1,"state":"merged"}]`, `["never break the wire format"]`)

	q := Query{Files: []string{"shared.go"}, Tests: []string{"shared_test.go"}}
	cr := NewContextResolver(0, nil) // default budget

	planner, _ := cr.Pack(ctx, store, q, RolePlanner)
	coder, _ := cr.Pack(ctx, store, q, RoleCoder)
	reviewer, _ := cr.Pack(ctx, store, q, RoleReviewer)
	tester, _ := cr.Pack(ctx, store, q, RoleTester)

	// Planner: files + decisions, but no tests and no PRs.
	if len(planner.RelevantFiles) == 0 || len(planner.RelevantTests) != 0 {
		t.Errorf("planner should have files but no tests: files=%v tests=%v", planner.RelevantFiles, planner.RelevantTests)
	}
	if len(planner.Decisions) == 0 {
		t.Error("planner should surface protected decisions")
	}
	if len(planner.RelatedTasks) > 0 && len(planner.RelatedTasks[0].PRs) != 0 {
		t.Error("planner should not include PRs")
	}

	// Coder: both files and tests.
	if len(coder.RelevantFiles) == 0 || len(coder.RelevantTests) == 0 {
		t.Errorf("coder should have both files and tests: files=%v tests=%v", coder.RelevantFiles, coder.RelevantTests)
	}

	// Reviewer: includes PRs and files.
	if len(reviewer.RelatedTasks) == 0 || len(reviewer.RelatedTasks[0].PRs) == 0 {
		t.Error("reviewer should include PRs")
	}
	if len(reviewer.RelevantFiles) == 0 {
		t.Error("reviewer should include affected files")
	}

	// Tester: tests but not files.
	if len(tester.RelevantTests) == 0 || len(tester.RelevantFiles) != 0 {
		t.Errorf("tester should have tests but no files: files=%v tests=%v", tester.RelevantFiles, tester.RelevantTests)
	}
}

func TestContextPackReviewerPrefersBugfixHistory(t *testing.T) {
	ctx := context.Background()
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()
	base := time.Date(2026, 7, 17, 8, 0, 0, 0, time.UTC)
	// The feature task has a stronger overlap (two shared files); the bugfix has
	// a weaker overlap (one shared file).
	seedRich(t, store, "FEAT", "feature", base, []string{"shared.go", "x.go"}, nil, "[]", "[]")
	seedRich(t, store, "BUG", "bugfix", base.Add(time.Hour), []string{"shared.go"}, nil, "[]", "[]")

	q := Query{Files: []string{"shared.go", "x.go"}}
	cr := NewContextResolver(0, nil)

	// Coder keeps overlap ranking: the feature (stronger overlap) leads.
	coder, _ := cr.Pack(ctx, store, q, RoleCoder)
	if coder.RelatedTasks[0].TaskID != "FEAT" {
		t.Fatalf("coder should rank by overlap; got %s first", coder.RelatedTasks[0].TaskID)
	}
	// Reviewer surfaces regression history first despite the weaker overlap.
	reviewer, _ := cr.Pack(ctx, store, q, RoleReviewer)
	if reviewer.RelatedTasks[0].TaskType != "bugfix" {
		t.Fatalf("reviewer should surface bugfix history first; got %+v", reviewer.RelatedTasks[0])
	}
}

func TestContextPackUnknownRoleErrors(t *testing.T) {
	store, err := memorydb.Open(t.TempDir(), "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()
	if _, err := NewContextResolver(0, nil).Pack(context.Background(), store, Query{}, Role("nonsense")); err == nil {
		t.Fatal("expected an error for an unknown role")
	}
}
