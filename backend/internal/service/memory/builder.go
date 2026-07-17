package memory

import (
	"context"
	"errors"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tinhtran24/maestro/backend/internal/adapters/memoryevents"
	"github.com/tinhtran24/maestro/backend/internal/domain"
)

// Differ reports the files a session branch changed, relative to where it forked
// from the repository's mainline. Implementations run against the shared
// repository so they keep working after the session worktree is removed.
type Differ interface {
	ChangedFiles(ctx context.Context, repoPath, branch string) ([]string, error)
}

// Builder turns a terminated session's durable facts into a completion event. It
// is best-effort about the diff: a failure to resolve changed files yields an
// event with empty file lists rather than an error, so capture never blocks on
// git state.
type Builder struct {
	diff  Differ
	now   func() time.Time
	newID func() string
	log   *slog.Logger
}

// NewBuilder returns a Builder using diff to resolve changed files.
func NewBuilder(diff Differ, log *slog.Logger) *Builder {
	if log == nil {
		log = slog.Default()
	}
	return &Builder{
		diff:  diff,
		now:   func() time.Time { return time.Now().UTC() },
		newID: func() string { return "EVT-" + uuid.NewString() },
		log:   log,
	}
}

// BuildCompletion assembles a task.completed event from a terminated session
// record. It gathers the intent, branch, kind, and harness from the record and
// derives changed files/tests from a diff of the session branch, splitting test
// files out. It returns an error only when the record cannot form a task (no id);
// diff failures degrade to empty file lists.
func (b *Builder) BuildCompletion(ctx context.Context, rec domain.SessionRecord, projectPath string) (memoryevents.Event, error) {
	if rec.ID == "" {
		return memoryevents.Event{}, errors.New("memory: session record has no id")
	}
	files, tests := b.changedFilesSplit(ctx, projectPath, rec.Metadata.Branch)
	return memoryevents.Event{
		V:          memoryevents.SchemaVersion,
		ID:         b.newID(),
		Type:       memoryevents.TypeTaskCompleted,
		OccurredAt: b.now(),
		Task: memoryevents.Task{
			ID:           string(rec.ID),
			SessionID:    string(rec.ID),
			ProjectID:    string(rec.ProjectID),
			Kind:         string(rec.Kind),
			Harness:      string(rec.Harness),
			Intent:       rec.Metadata.Prompt,
			TaskType:     classifyTaskType(rec),
			Branch:       rec.Metadata.Branch,
			ChangedFiles: files,
			ChangedTests: tests,
		},
	}, nil
}

// changedFilesSplit resolves the branch's changed files and partitions them into
// non-test files and test files. A missing branch or a diff failure yields two
// empty slices; the failure is logged at debug, never surfaced.
func (b *Builder) changedFilesSplit(ctx context.Context, projectPath, branch string) (files, tests []string) {
	if projectPath == "" || branch == "" || b.diff == nil {
		return nil, nil
	}
	changed, err := b.diff.ChangedFiles(ctx, projectPath, branch)
	if err != nil {
		b.log.Debug("memory: changed-file diff failed; recording empty file lists", "branch", branch, "error", err)
		return nil, nil
	}
	for _, f := range changed {
		if f == "" {
			continue
		}
		if isTestFile(f) {
			tests = append(tests, f)
		} else {
			files = append(files, f)
		}
	}
	return files, tests
}

// isTestFile recognises common cross-language test-file conventions so test
// changes can be weighted separately from source changes.
func isTestFile(p string) bool {
	lower := strings.ToLower(p)
	base := path.Base(lower)
	switch {
	case strings.HasSuffix(base, "_test.go"): // Go
		return true
	case strings.HasSuffix(base, "_test.py"), strings.HasPrefix(base, "test_"): // Python
		return true
	case strings.HasSuffix(base, "_spec.rb"), strings.HasSuffix(base, "_test.rb"): // Ruby
		return true
	case strings.Contains(base, ".test."), strings.Contains(base, ".spec."): // JS/TS
		return true
	}
	for _, seg := range strings.Split(lower, "/") {
		switch seg {
		case "test", "tests", "__tests__", "spec", "specs":
			return true
		}
	}
	return false
}

// classifyTaskType infers a coarse task type from the conventional-commit prefix
// on the branch, suggested branch, or commit message. It returns "" when nothing
// matches; the type is a heuristic that later phases may refine.
func classifyTaskType(rec domain.SessionRecord) string {
	for _, s := range []string{rec.Metadata.Branch, rec.Metadata.SuggestedBranch, rec.Metadata.CommitMessage} {
		if t := taskTypeFromPrefix(s); t != "" {
			return t
		}
	}
	return ""
}

func taskTypeFromPrefix(s string) string {
	prefix := strings.ToLower(strings.TrimSpace(s))
	for _, sep := range []string{"/", ":", "(", " "} {
		if i := strings.Index(prefix, sep); i >= 0 {
			prefix = prefix[:i]
		}
	}
	switch strings.TrimSpace(prefix) {
	case "feat", "feature", "feats":
		return "feature"
	case "fix", "bugfix", "hotfix", "bug":
		return "bugfix"
	case "refactor":
		return "refactor"
	case "docs", "doc":
		return "docs"
	case "test", "tests":
		return "test"
	case "chore":
		return "chore"
	case "perf":
		return "perf"
	case "ci":
		return "ci"
	case "build":
		return "build"
	default:
		return ""
	}
}
