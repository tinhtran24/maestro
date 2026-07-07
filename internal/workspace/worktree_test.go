package workspace

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordingRunner struct {
	failShowRef bool
	outputs     map[string]string
	calls       []string
}

func (r *recordingRunner) Run(_ context.Context, _ string, command string, args ...string) error {
	r.calls = append(r.calls, command+" "+strings.Join(args, " "))
	if command == "git" && len(args) >= 1 && args[0] == "show-ref" && r.failShowRef {
		return os.ErrNotExist
	}
	return nil
}

func (r *recordingRunner) Output(_ context.Context, _ string, command string, args ...string) (string, error) {
	call := command + " " + strings.Join(args, " ")
	r.calls = append(r.calls, call)
	if r.outputs != nil {
		return r.outputs[call], nil
	}
	return "", nil
}

func TestWorktreeManagerCreatesBranchWorktreeForTask(t *testing.T) {
	root := t.TempDir()
	runner := &recordingRunner{
		failShowRef: true,
		outputs: map[string]string{
			"git rev-parse HEAD": "abc123\n",
		},
	}
	manager := WorktreeManager{Workspace: Open(root), Runner: runner}

	spec, err := manager.Prepare(context.Background(), WorktreeSpec{
		TaskID: "T-106",
		Branch: "thanos/T-106-cart",
	})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Path != filepath.Join(root, ".thanos", "worktrees", "T-106") {
		t.Fatalf("path = %s", spec.Path)
	}
	want := "git worktree add -b thanos/T-106-cart " + spec.Path + " HEAD"
	if got := runner.calls[len(runner.calls)-1]; got != want {
		t.Fatalf("last call = %q, want %q", got, want)
	}
	if spec.BaseCommit != "abc123" {
		t.Fatalf("base commit = %q", spec.BaseCommit)
	}
	data, err := os.ReadFile(filepath.Join(spec.Path, ".thanos-task.json"))
	if err != nil {
		t.Fatalf("manifest was not written: %v", err)
	}
	var manifest TaskRuntimeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("manifest json: %v", err)
	}
	if manifest.TaskID != "T-106" || manifest.Branch != "thanos/T-106-cart" || manifest.BaseCommit != "abc123" {
		t.Fatalf("manifest = %#v", manifest)
	}
}

func TestWorktreeManagerRefusesProtectedBranch(t *testing.T) {
	manager := WorktreeManager{Workspace: Open(t.TempDir()), Runner: &recordingRunner{}}
	_, err := manager.Prepare(context.Background(), WorktreeSpec{TaskID: "T-1", Branch: "main"})
	if err == nil || !strings.Contains(err.Error(), "protected branch") {
		t.Fatalf("expected protected branch error, got %v", err)
	}
}

func TestWorktreeManagerUsesExistingBranch(t *testing.T) {
	root := t.TempDir()
	runner := &recordingRunner{}
	manager := WorktreeManager{Workspace: Open(root), Runner: runner}
	spec, err := manager.Prepare(context.Background(), WorktreeSpec{TaskID: "T-2", Branch: "thanos/T-2"})
	if err != nil {
		t.Fatal(err)
	}
	want := "git worktree add " + spec.Path + " thanos/T-2"
	if got := runner.calls[len(runner.calls)-1]; got != want {
		t.Fatalf("last call = %q, want %q", got, want)
	}
}

func TestWorktreeManagerSyncDiffAndCleanup(t *testing.T) {
	root := t.TempDir()
	runner := &recordingRunner{outputs: map[string]string{
		"git diff origin/main...HEAD": "diff --git a/file b/file\n",
	}}
	manager := WorktreeManager{Workspace: Open(root), Runner: runner}
	spec := WorktreeSpec{TaskID: "T-3", Branch: "thanos/T-3", Path: filepath.Join(root, ".thanos", "worktrees", "T-3")}

	if err := manager.Sync(context.Background(), spec, "origin/main"); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	diff, err := manager.Diff(context.Background(), spec, "origin/main")
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if diff != "diff --git a/file b/file\n" {
		t.Fatalf("diff = %q", diff)
	}
	if err := manager.Cleanup(context.Background(), spec, false); err == nil || !strings.Contains(err.Error(), "confirmation") {
		t.Fatalf("expected confirmation error, got %v", err)
	}
	if err := manager.Cleanup(context.Background(), spec, true); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	joined := strings.Join(runner.calls, "\n")
	for _, want := range []string{
		"git rebase origin/main",
		"git diff origin/main...HEAD",
		"git worktree remove " + spec.Path,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("calls missing %q:\n%s", want, joined)
		}
	}
}
