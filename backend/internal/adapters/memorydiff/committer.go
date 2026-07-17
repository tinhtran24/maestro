package memorydiff

import "context"

// Committer commits a single file into a git repository. It is used to persist
// the appended project-memory event log when a project opts into auto-commit.
// It never pushes and never switches branches: the commit lands on the
// repository's current branch, and sharing across machines is the user's normal
// git flow.
type Committer struct {
	binary string
}

// NewCommitter returns a Committer that invokes the git binary on PATH.
func NewCommitter() *Committer { return &Committer{binary: "git"} }

// CommitFile stages and commits exactly relPath (a repo-relative pathspec) in
// the repository at repoPath, leaving every other working-tree and staged change
// untouched. The pathspec on both add and commit scopes the commit to that one
// file, so a user's in-progress work is never swept into the memory commit.
func (c *Committer) CommitFile(ctx context.Context, repoPath, relPath, message string) error {
	if _, err := runGit(ctx, c.binary, "-C", repoPath, "add", "--", relPath); err != nil {
		return err
	}
	if _, err := runGit(ctx, c.binary, "-C", repoPath, "commit", "-m", message, "--", relPath); err != nil {
		return err
	}
	return nil
}
