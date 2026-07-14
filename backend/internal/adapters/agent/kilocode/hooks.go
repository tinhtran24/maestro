package kilocode

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/tinhtran24/maestro/backend/internal/adapters/agent/hookutil"
	"github.com/tinhtran24/maestro/backend/internal/ports"
)

const (
	// Kilo Code scans each config dir for `{plugin,plugins}/*.{ts,js}` (verified
	// in the @kilocode/cli binary). Its config-dir suffixes are `.kilo`,
	// `.kilocode`, and `.opencode` (it is an opencode fork). Maestro writes the
	// branded `.kilocode/plugins/` so the Maestro plugin lands in Kilo's own dir and
	// never collides with a sibling opencode adapter's `.opencode/` install.
	kilocodePluginDirName = ".kilocode"
	kilocodePluginSubDir  = "plugins"

	// kilocodePluginFileName is the Maestro-owned plugin file. Maestro fully owns this
	// filename: install overwrites it and uninstall deletes it (guarded by the
	// sentinel), so user-authored plugins in other files are never touched.
	// It is TypeScript (Kilo runs on Bun); the file's only import is a type-only
	// import, which Bun erases at runtime.
	kilocodePluginFileName = "maestro-activity.ts"

	// kilocodePluginSentinel marks the file as Maestro-managed. AreHooksInstalled and
	// UninstallHooks key off it so Maestro never deletes a user file that happens to
	// share the name. It must appear verbatim in the embedded plugin source.
	kilocodePluginSentinel = "maestro: managed kilocode activity plugin"

	// kilocodeHookCommandPrefix identifies the hook commands Maestro owns. The
	// embedded plugin shells `maestro hooks kilocode <event>`; this prefix is the
	// shared contract with the `maestro hooks` CLI dispatcher and is asserted by tests
	// so the plugin can't silently drift away from it.
	kilocodeHookCommandPrefix = "maestro hooks kilocode "
)

// kilocodePluginSource is the Maestro-managed Kilo Code plugin, embedded so it ships
// inside the binary and is written verbatim into a session's worktree on hook
// install. It is a real, lintable source file under assets/ rather than a Go
// string literal because it is plugin source code, not a data structure Maestro
// assembles (the way it builds Codex/Claude hook JSON).
//
//go:embed assets/maestro-activity.ts
var kilocodePluginSource string

// kilocodeManagedEvents are the normalized activity events the embedded plugin
// reports. They are defined here (not parsed from the file) so tests can assert
// the plugin wires every one via the `maestro hooks kilocode <event>` command, and
// they mirror exactly the events kilocode.DeriveActivityState switches on.
var kilocodeManagedEvents = []string{"session-start", "user-prompt-submit", "permission-request", "stop"}

// GetAgentHooks installs Maestro's Kilo Code activity plugin into the worktree-local
// .kilocode/plugins/ directory. Unlike Claude Code and Codex, Kilo Code has no
// native command-hook config to merge into; its only lifecycle-extensibility
// surface is a JS/TS plugin. Maestro therefore writes a dedicated, Maestro-owned plugin
// file. The write is atomic and idempotent: re-installing overwrites Maestro's own
// file with identical content. It refuses to overwrite a file that is NOT
// Maestro-managed (no sentinel), so a user plugin that happens to occupy our path is
// never silently destroyed — install fails loudly instead.
func (p *Plugin) GetAgentHooks(ctx context.Context, cfg ports.WorkspaceHookConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.WorkspacePath) == "" {
		return errors.New("kilocode.GetAgentHooks: WorkspacePath is required")
	}

	pluginPath := kilocodePluginPath(cfg.WorkspacePath)
	// Guard against clobbering a user file at our path: overwrite only when the
	// target is absent or already Maestro-managed. A foreign file is a loud error,
	// not silent data loss (uninstall is sentinel-guarded the same way).
	if _, err := os.Stat(pluginPath); err == nil {
		managed, err := isAOManagedPlugin(pluginPath)
		if err != nil {
			return fmt.Errorf("kilocode.GetAgentHooks: %w", err)
		}
		if !managed {
			return fmt.Errorf("kilocode.GetAgentHooks: refusing to overwrite non-Maestro file at %s — move it so Maestro can install its plugin", pluginPath)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("kilocode.GetAgentHooks: stat plugin: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(pluginPath), 0o750); err != nil {
		return fmt.Errorf("kilocode.GetAgentHooks: create plugin dir: %w", err)
	}
	if err := hookutil.AtomicWriteFile(pluginPath, []byte(kilocodePluginSource), 0o600); err != nil {
		return fmt.Errorf("kilocode.GetAgentHooks: write plugin: %w", err)
	}
	if err := hookutil.EnsureWorkspaceGitignore(filepath.Dir(pluginPath), kilocodePluginFileName); err != nil {
		return fmt.Errorf("kilocode.GetAgentHooks: gitignore: %w", err)
	}
	return nil
}

// UninstallHooks removes Maestro's Kilo Code plugin from the workspace-local
// .kilocode/plugins/ directory. It deletes the file only when it carries the Maestro
// sentinel, so a user file that happens to share the name is left in place. A
// missing file is a no-op.
func (p *Plugin) UninstallHooks(ctx context.Context, workspacePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(workspacePath) == "" {
		return errors.New("kilocode.UninstallHooks: workspacePath is required")
	}

	pluginPath := kilocodePluginPath(workspacePath)
	managed, err := isAOManagedPlugin(pluginPath)
	if err != nil {
		return fmt.Errorf("kilocode.UninstallHooks: %w", err)
	}
	if !managed {
		return nil
	}
	if err := os.Remove(pluginPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("kilocode.UninstallHooks: remove plugin: %w", err)
	}
	return nil
}

// AreHooksInstalled reports whether Maestro's Kilo Code plugin is present in the
// workspace-local plugin dir. A missing file, or a same-named file without the
// Maestro sentinel, means none are installed.
func (p *Plugin) AreHooksInstalled(ctx context.Context, workspacePath string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if strings.TrimSpace(workspacePath) == "" {
		return false, errors.New("kilocode.AreHooksInstalled: workspacePath is required")
	}
	managed, err := isAOManagedPlugin(kilocodePluginPath(workspacePath))
	if err != nil {
		return false, fmt.Errorf("kilocode.AreHooksInstalled: %w", err)
	}
	return managed, nil
}

func kilocodePluginPath(workspacePath string) string {
	return filepath.Join(workspacePath, kilocodePluginDirName, kilocodePluginSubDir, kilocodePluginFileName)
}

// isAOManagedPlugin reports whether the file at path exists and carries the Maestro
// sentinel. A missing file yields (false, nil).
func isAOManagedPlugin(path string) (bool, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path built from caller-owned workspace dir
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	return strings.Contains(string(data), kilocodePluginSentinel), nil
}
