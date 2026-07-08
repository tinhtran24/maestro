package qwen

import (
	"context"
	"path/filepath"

	"github.com/tinhtran/thanos/backend/internal/adapters/agent/hooksjson"
	"github.com/tinhtran/thanos/backend/internal/ports"
)

const (
	qwenSettingsDirName  = ".qwen"
	qwenSettingsFileName = "settings.json"

	// qwenHookCommandPrefix identifies the hook commands Thanos owns, so install
	// skips duplicates and uninstall recognizes Thanos entries by prefix.
	qwenHookCommandPrefix = "to hooks qwen "

	// qwenHookTimeout is in milliseconds: Qwen Code (a gemini-cli fork) measures
	// hook timeouts in ms, unlike Claude/Codex which use seconds.
	qwenHookTimeout = 30000
)

// qwenStartupMatcher is referenced by pointer so SessionStart serializes with
// its "startup" source matcher.
var qwenStartupMatcher = "startup"

// qwenManagedHooks is the source of truth for the hooks Thanos installs:
// SessionStart (under the "startup" source matcher), UserPromptSubmit,
// PermissionRequest, and Stop.
var qwenManagedHooks = []hooksjson.HookSpec{
	{Event: "SessionStart", Matcher: &qwenStartupMatcher, Command: qwenHookCommandPrefix + "session-start"},
	{Event: "UserPromptSubmit", Command: qwenHookCommandPrefix + "user-prompt-submit"},
	{Event: "PermissionRequest", Command: qwenHookCommandPrefix + "permission-request"},
	{Event: "Stop", Command: qwenHookCommandPrefix + "stop"},
}

// qwenHooks manages Thanos's hooks in the workspace-local .qwen/settings.json file.
var qwenHooks = hooksjson.Manager{
	Label:         "qwen",
	CommandPrefix: qwenHookCommandPrefix,
	Timeout:       qwenHookTimeout,
	Path:          qwenSettingsPath,
	Managed:       qwenManagedHooks,
}

func qwenSettingsPath(workspacePath string) string {
	return filepath.Join(workspacePath, qwenSettingsDirName, qwenSettingsFileName)
}

// GetAgentHooks installs Thanos's Qwen Code hooks, preserving user-defined hooks and unrelated settings.
func (p *Plugin) GetAgentHooks(ctx context.Context, cfg ports.WorkspaceHookConfig) error {
	return qwenHooks.Install(ctx, cfg.WorkspacePath)
}

// UninstallHooks removes Thanos's Qwen Code hooks, leaving user-defined hooks untouched.
func (p *Plugin) UninstallHooks(ctx context.Context, workspacePath string) error {
	return qwenHooks.Uninstall(ctx, workspacePath)
}

// AreHooksInstalled reports whether any Thanos Qwen Code hook is present.
func (p *Plugin) AreHooksInstalled(ctx context.Context, workspacePath string) (bool, error) {
	return qwenHooks.AreInstalled(ctx, workspacePath)
}
