package memory

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memorydiff"
	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/domain"
	"github.com/tinhtran24/maestro/backend/internal/storage/memorydb"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput() // #nosec G204 -- test-only, fixed git args
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init")
	git(t, dir, "config", "user.email", "test@example.com")
	git(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x\n"), 0o600); err != nil {
		t.Fatalf("seed readme: %v", err)
	}
	git(t, dir, "add", "README.md")
	git(t, dir, "commit", "-m", "init")
	return dir
}

func commitCount(t *testing.T, repo string) int {
	t.Helper()
	n, err := strconv.Atoi(strings.TrimSpace(git(t, repo, "rev-list", "--count", "HEAD")))
	if err != nil {
		t.Fatalf("parse commit count: %v", err)
	}
	return n
}

func autoCommitRecorder(t *testing.T, cacheDir string, projects ProjectLookup) *Recorder {
	t.Helper()
	events := memoryevents.New()
	builder := NewBuilder(&fakeDiffer{}, nil)
	projector := NewProjector(events, nil)
	return NewRecorder(builder, events, projector, projects, memorydiff.NewCommitter(), cacheDir, nil)
}

func worker(id string) domain.SessionRecord {
	return domain.SessionRecord{ID: domain.SessionID(id), ProjectID: "acme", Kind: domain.KindWorker, Metadata: domain.SessionMetadata{Prompt: "do it"}}
}

func TestAutoCommitOffLeavesEventsUncommitted(t *testing.T) {
	repo := initGitRepo(t)
	projects := fakeProjects{rec: domain.ProjectRecord{ID: "acme", Path: repo}, ok: true} // AutoCommit defaults off
	r := autoCommitRecorder(t, t.TempDir(), projects)

	before := commitCount(t, repo)
	r.RecordCompletion(context.Background(), worker("acme-1"))

	if _, err := os.Stat(filepath.Join(repo, ".maestro", "memory", "events.jsonl")); err != nil {
		t.Fatalf("events.jsonl should exist: %v", err)
	}
	if after := commitCount(t, repo); after != before {
		t.Fatalf("auto-commit off must not commit: commits %d -> %d", before, after)
	}
	if status := git(t, repo, "status", "--porcelain", "-uall"); !strings.Contains(status, ".maestro/memory/events.jsonl") {
		t.Fatalf("events.jsonl should be an uncommitted working-tree change, status:\n%s", status)
	}
}

func TestAutoCommitOnCommitsOnlyEventsAndNeverPushes(t *testing.T) {
	repo := initGitRepo(t)
	projects := fakeProjects{rec: domain.ProjectRecord{
		ID: "acme", Path: repo, Config: domain.ProjectConfig{Memory: domain.MemoryConfig{AutoCommit: true}},
	}, ok: true}
	r := autoCommitRecorder(t, t.TempDir(), projects)

	// An unrelated dirty file must survive untouched: the memory commit is scoped.
	if err := os.WriteFile(filepath.Join(repo, "unrelated.go"), []byte("package x\n"), 0o600); err != nil {
		t.Fatalf("write unrelated: %v", err)
	}

	before := commitCount(t, repo)
	r.RecordCompletion(context.Background(), worker("acme-1"))

	if after := commitCount(t, repo); after != before+1 {
		t.Fatalf("auto-commit on must add exactly one commit: %d -> %d", before, after)
	}
	// The commit touches only events.jsonl.
	touched := git(t, repo, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(touched, ".maestro/memory/events.jsonl") {
		t.Fatalf("memory commit should include events.jsonl, touched:\n%s", touched)
	}
	if strings.Contains(touched, "unrelated.go") {
		t.Fatalf("memory commit must not sweep in the user's other work, touched:\n%s", touched)
	}
	// The unrelated file stays uncommitted; events.jsonl is now clean.
	status := git(t, repo, "status", "--porcelain")
	if !strings.Contains(status, "unrelated.go") {
		t.Fatalf("unrelated work should remain uncommitted, status:\n%s", status)
	}
	if strings.Contains(status, ".maestro/memory/events.jsonl") {
		t.Fatalf("events.jsonl should be committed and clean, status:\n%s", status)
	}
	// No push: the repo has no remote, and the committer never adds one.
	if remotes := strings.TrimSpace(git(t, repo, "remote")); remotes != "" {
		t.Fatalf("auto-commit must not configure or push to a remote, got remotes: %q", remotes)
	}
}

func TestProjectionDBLivesOutsideTheProjectRepo(t *testing.T) {
	repo := initGitRepo(t)
	cacheDir := t.TempDir()
	store, err := memorydb.Open(cacheDir, "acme")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = store.Close() }()
	if strings.HasPrefix(store.Path(), repo) {
		t.Fatalf("projection DB %q must live outside the project repo %q", store.Path(), repo)
	}
	if !strings.HasPrefix(store.Path(), cacheDir) {
		t.Fatalf("projection DB %q must live under the app-state cache %q", store.Path(), cacheDir)
	}
}
