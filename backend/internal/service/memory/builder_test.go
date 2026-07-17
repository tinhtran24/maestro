package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/tinhtran24/maestro/backend/internal/domain"
)

type fakeDiffer struct {
	files  []string
	err    error
	calls  int
	branch string
}

func (f *fakeDiffer) ChangedFiles(_ context.Context, _, branch string) ([]string, error) {
	f.calls++
	f.branch = branch
	return f.files, f.err
}

func TestBuildCompletionSplitsTestsFromFiles(t *testing.T) {
	diff := &fakeDiffer{files: []string{
		"internal/widget.go",
		"internal/widget_test.go",
		"web/src/x.ts",
		"web/src/x.spec.ts",
		"py/test_y.py",
		"py/y.py",
		"tests/helper.go",
	}}
	b := NewBuilder(diff, nil)
	rec := domain.SessionRecord{
		ID:        "acme-1",
		ProjectID: "acme",
		Kind:      domain.KindWorker,
		Metadata:  domain.SessionMetadata{Branch: "feat/widget", Prompt: "add the widget"},
	}

	ev, err := b.BuildCompletion(context.Background(), rec, "/repo")
	if err != nil {
		t.Fatalf("BuildCompletion: %v", err)
	}
	wantFiles := []string{"internal/widget.go", "web/src/x.ts", "py/y.py"}
	wantTests := []string{"internal/widget_test.go", "web/src/x.spec.ts", "py/test_y.py", "tests/helper.go"}
	if !reflect.DeepEqual(ev.Task.ChangedFiles, wantFiles) {
		t.Errorf("changed files = %v, want %v", ev.Task.ChangedFiles, wantFiles)
	}
	if !reflect.DeepEqual(ev.Task.ChangedTests, wantTests) {
		t.Errorf("changed tests = %v, want %v", ev.Task.ChangedTests, wantTests)
	}
	if ev.Task.TaskType != "feature" {
		t.Errorf("task type = %q, want feature", ev.Task.TaskType)
	}
	if ev.Task.Intent != "add the widget" || ev.Task.SessionID != "acme-1" || ev.Task.Kind != "worker" {
		t.Errorf("task fields not carried through: %+v", ev.Task)
	}
	if ev.ID == "" || ev.Type != "task.completed" || ev.OccurredAt.IsZero() {
		t.Errorf("event envelope incomplete: id=%q type=%q at=%v", ev.ID, ev.Type, ev.OccurredAt)
	}
}

func TestBuildCompletionEmptyBranchSkipsDiff(t *testing.T) {
	diff := &fakeDiffer{files: []string{"a.go"}}
	b := NewBuilder(diff, nil)
	rec := domain.SessionRecord{ID: "acme-1", ProjectID: "acme", Kind: domain.KindWorker}
	ev, err := b.BuildCompletion(context.Background(), rec, "/repo")
	if err != nil {
		t.Fatalf("BuildCompletion: %v", err)
	}
	if diff.calls != 0 {
		t.Errorf("differ should not be called without a branch, calls=%d", diff.calls)
	}
	if len(ev.Task.ChangedFiles) != 0 || len(ev.Task.ChangedTests) != 0 {
		t.Errorf("expected no files without a branch, got %+v", ev.Task)
	}
}

func TestBuildCompletionDiffErrorYieldsEmptyLists(t *testing.T) {
	b := NewBuilder(&fakeDiffer{err: errors.New("branch gone")}, nil)
	rec := domain.SessionRecord{ID: "acme-1", ProjectID: "acme", Metadata: domain.SessionMetadata{Branch: "feat/x"}}
	ev, err := b.BuildCompletion(context.Background(), rec, "/repo")
	if err != nil {
		t.Fatalf("diff failure must not error, got %v", err)
	}
	if len(ev.Task.ChangedFiles) != 0 || len(ev.Task.ChangedTests) != 0 {
		t.Errorf("expected empty lists on diff error, got %+v", ev.Task)
	}
}

func TestBuildCompletionRejectsMissingID(t *testing.T) {
	b := NewBuilder(&fakeDiffer{}, nil)
	if _, err := b.BuildCompletion(context.Background(), domain.SessionRecord{}, "/repo"); err == nil {
		t.Fatal("expected an error for a record with no id")
	}
}

func TestClassifyTaskType(t *testing.T) {
	cases := []struct {
		branch, suggested, commit, want string
	}{
		{branch: "feat/x", want: "feature"},
		{branch: "feature/x", want: "feature"},
		{branch: "fix/x", want: "bugfix"},
		{branch: "hotfix/y", want: "bugfix"},
		{branch: "refactor/z", want: "refactor"},
		{branch: "docs/z", want: "docs"},
		{branch: "chore/z", want: "chore"},
		{branch: "", commit: "fix: correct the thing", want: "bugfix"},
		{branch: "", suggested: "feat/from-suggested", want: "feature"},
		{branch: "random-branch", want: ""},
	}
	for _, c := range cases {
		rec := domain.SessionRecord{Metadata: domain.SessionMetadata{
			Branch: c.branch, SuggestedBranch: c.suggested, CommitMessage: c.commit,
		}}
		if got := classifyTaskType(rec); got != c.want {
			t.Errorf("classify(branch=%q suggested=%q commit=%q) = %q, want %q", c.branch, c.suggested, c.commit, got, c.want)
		}
	}
}

func TestIsTestFile(t *testing.T) {
	tests := map[string]bool{
		"internal/foo_test.go": true,
		"internal/foo.go":      false,
		"web/x.spec.ts":        true,
		"web/x.test.tsx":       true,
		"web/x.ts":             false,
		"py/test_thing.py":     true,
		"py/thing.py":          false,
		"tests/whatever.go":    true,
		"src/__tests__/a.js":   true,
		"spec/models/user.rb":  true,
		"README.md":            false,
	}
	for path, want := range tests {
		if got := isTestFile(path); got != want {
			t.Errorf("isTestFile(%q) = %v, want %v", path, got, want)
		}
	}
}
