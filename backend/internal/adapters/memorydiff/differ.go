// Package memorydiff resolves the files a session branch changed, for project
// memory capture. It shells out to git against the shared repository, so it
// keeps working after the session worktree has been removed: the branch ref and
// its objects still live in the project's primary checkout.
package memorydiff

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	aoprocess "github.com/tinhtran24/maestro/backend/internal/process"
)

// Differ computes changed-file lists via git.
type Differ struct {
	binary string
}

// New returns a Differ that invokes the git binary on PATH.
func New() *Differ { return &Differ{binary: "git"} }

// ChangedFiles returns the files branch changed relative to its fork point from
// the repository's current HEAD (typically the checked-out default branch), run
// inside repoPath. It returns nil (not an error) when the branch cannot be
// diffed (already merged and deleted, unrelated history, no merge base) so
// callers treating capture as best-effort record empty file lists.
func (d *Differ) ChangedFiles(ctx context.Context, repoPath, branch string) ([]string, error) {
	if repoPath == "" || branch == "" {
		return nil, nil
	}
	base, err := d.run(ctx, "-C", repoPath, "merge-base", branch, "HEAD")
	if err != nil {
		return nil, nil //nolint:nilerr // no merge base (branch merged/deleted/unrelated) means no capturable diff; best-effort
	}
	base = strings.TrimSpace(base)
	if base == "" {
		return nil, nil
	}
	out, err := d.run(ctx, "-C", repoPath, "diff", "--name-only", base, branch)
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

func (d *Differ) run(ctx context.Context, args ...string) (string, error) {
	return runGit(ctx, d.binary, args...)
}

// runGit executes git and returns stdout, wrapping a failure with stderr.
func runGit(ctx context.Context, binary string, args ...string) (string, error) {
	cmd := aoprocess.CommandContext(ctx, binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func splitLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			out = append(out, t)
		}
	}
	return out
}
