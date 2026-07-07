package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type CommandRunner interface {
	Run(ctx context.Context, dir string, command string, args ...string) error
}

type OutputCommandRunner interface {
	CommandRunner
	Output(ctx context.Context, dir string, command string, args ...string) (string, error)
}

type ExecCommandRunner struct{}

func (ExecCommandRunner) Run(ctx context.Context, dir string, command string, args ...string) error {
	_, err := (ExecCommandRunner{}).Output(ctx, dir, command, args...)
	return err
}

func (ExecCommandRunner) Output(ctx context.Context, dir string, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%s %s: %w: %s", command, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

type WorktreeSpec struct {
	TaskID     string `json:"task_id"`
	Branch     string `json:"branch"`
	Path       string `json:"path"`
	Base       string `json:"base"`
	BaseCommit string `json:"base_commit,omitempty"`
}

type TaskRuntimeManifest struct {
	SchemaVersion int       `json:"schema_version"`
	TaskID        string    `json:"task_id"`
	Branch        string    `json:"branch"`
	WorktreePath  string    `json:"worktree_path"`
	Base          string    `json:"base"`
	BaseCommit    string    `json:"base_commit"`
	WorkspaceRoot string    `json:"workspace_root"`
	CreatedAt     time.Time `json:"created_at"`
}

type WorktreeManager struct {
	Workspace *Workspace
	Runner    CommandRunner
}

func (m WorktreeManager) Prepare(ctx context.Context, spec WorktreeSpec) (WorktreeSpec, error) {
	if m.Workspace == nil {
		return spec, fmt.Errorf("workspace is required")
	}
	if m.Runner == nil {
		return spec, fmt.Errorf("command runner is required")
	}
	if strings.TrimSpace(spec.TaskID) == "" {
		return spec, fmt.Errorf("task id is required")
	}
	if strings.TrimSpace(spec.Branch) == "" {
		return spec, fmt.Errorf("branch is required")
	}
	if isMainBranch(spec.Branch) {
		return spec, fmt.Errorf("refusing to use protected branch %q for task worktree", spec.Branch)
	}
	if strings.TrimSpace(spec.Base) == "" {
		spec.Base = "HEAD"
	}
	if strings.TrimSpace(spec.Path) == "" {
		spec.Path = m.Workspace.TaskWorktreePath(spec.TaskID)
	}
	if err := os.MkdirAll(filepath.Dir(spec.Path), 0o755); err != nil {
		return spec, err
	}
	if _, err := os.Stat(spec.Path); err == nil {
		return spec, nil
	} else if !os.IsNotExist(err) {
		return spec, err
	}
	if err := m.Runner.Run(ctx, m.Workspace.Root, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return spec, fmt.Errorf("workspace is not a git repository: %w", err)
	}
	if output, ok := m.Runner.(OutputCommandRunner); ok {
		baseCommit, err := output.Output(ctx, m.Workspace.Root, "git", "rev-parse", spec.Base)
		if err != nil {
			return spec, fmt.Errorf("resolve base commit: %w", err)
		}
		spec.BaseCommit = strings.TrimSpace(baseCommit)
	}
	if err := m.Runner.Run(ctx, m.Workspace.Root, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+spec.Branch); err == nil {
		if err := m.Runner.Run(ctx, m.Workspace.Root, "git", "worktree", "add", spec.Path, spec.Branch); err != nil {
			return spec, err
		}
		return spec, m.WriteManifest(spec)
	}
	if err := m.Runner.Run(ctx, m.Workspace.Root, "git", "worktree", "add", "-b", spec.Branch, spec.Path, spec.Base); err != nil {
		return spec, err
	}
	return spec, m.WriteManifest(spec)
}

func (m WorktreeManager) WriteManifest(spec WorktreeSpec) error {
	if m.Workspace == nil {
		return fmt.Errorf("workspace is required")
	}
	if strings.TrimSpace(spec.Path) == "" {
		return fmt.Errorf("worktree path is required")
	}
	manifest := TaskRuntimeManifest{
		SchemaVersion: 1,
		TaskID:        spec.TaskID,
		Branch:        spec.Branch,
		WorktreePath:  filepath.ToSlash(spec.Path),
		Base:          spec.Base,
		BaseCommit:    spec.BaseCommit,
		WorkspaceRoot: m.Workspace.Root,
		CreatedAt:     time.Now().UTC(),
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(spec.Path, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(spec.Path, ".thanos-task.json"), append(data, '\n'), 0o644)
}

func (m WorktreeManager) Sync(ctx context.Context, spec WorktreeSpec, defaultBranch string) error {
	if m.Runner == nil {
		return fmt.Errorf("command runner is required")
	}
	if strings.TrimSpace(spec.Path) == "" {
		return fmt.Errorf("worktree path is required")
	}
	defaultBranch = strings.TrimSpace(defaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	return m.Runner.Run(ctx, spec.Path, "git", "rebase", defaultBranch)
}

func (m WorktreeManager) Diff(ctx context.Context, spec WorktreeSpec, defaultBranch string) (string, error) {
	runner, ok := m.Runner.(OutputCommandRunner)
	if !ok {
		return "", fmt.Errorf("command runner does not support output capture")
	}
	if strings.TrimSpace(spec.Path) == "" {
		return "", fmt.Errorf("worktree path is required")
	}
	defaultBranch = strings.TrimSpace(defaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	return runner.Output(ctx, spec.Path, "git", "diff", defaultBranch+"...HEAD")
}

func (m WorktreeManager) Cleanup(ctx context.Context, spec WorktreeSpec, confirmed bool) error {
	if !confirmed {
		return fmt.Errorf("worktree cleanup requires explicit confirmation")
	}
	if m.Workspace == nil {
		return fmt.Errorf("workspace is required")
	}
	if m.Runner == nil {
		return fmt.Errorf("command runner is required")
	}
	if strings.TrimSpace(spec.Path) == "" {
		spec.Path = m.Workspace.TaskWorktreePath(spec.TaskID)
	}
	if strings.TrimSpace(spec.Path) == "" {
		return fmt.Errorf("worktree path is required")
	}
	return m.Runner.Run(ctx, m.Workspace.Root, "git", "worktree", "remove", spec.Path)
}

func isMainBranch(branch string) bool {
	switch strings.TrimSpace(branch) {
	case "main", "master", "trunk":
		return true
	default:
		return false
	}
}
