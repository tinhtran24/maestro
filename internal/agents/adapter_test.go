package agents

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestDefaultAdapterRegistryContainsRequiredAdapters(t *testing.T) {
	registry := DefaultAdapterRegistry()
	for _, id := range []string{"claude-code", "codex", "gemini-cli", "opencode", "shell"} {
		if _, ok := registry.Get(id); !ok {
			t.Fatalf("missing adapter %q", id)
		}
	}
	if adapter, ok := registry.Get("claude"); !ok || adapter.ID() != "claude-code" {
		t.Fatalf("claude alias = %v/%v, want claude-code", adapter, ok)
	}
	if adapter, ok := registry.Get("gemini"); !ok || adapter.ID() != "gemini-cli" {
		t.Fatalf("gemini alias = %v/%v, want gemini-cli", adapter, ok)
	}
}

func TestDetectAllUsesCheapLocalDetection(t *testing.T) {
	registry := DefaultAdapterRegistry()
	detection := DetectionContext{
		LookPath: func(command string) (string, error) {
			if command == "codex" || command == "claude" {
				return "/usr/local/bin/" + command, nil
			}
			return "", errors.New("not found")
		},
		Version: func(command string) (string, error) {
			if command == "codex" {
				return "1.2.3", nil
			}
			return "", nil
		},
		Auth: func(providerID string) (AuthStatus, error) {
			if providerID == "claude-code" {
				return AuthUnauthorized, nil
			}
			return AuthUnknown, nil
		},
	}
	providers, err := registry.DetectAll(context.Background(), detection)
	if err != nil {
		t.Fatalf("DetectAll: %v", err)
	}
	byID := map[string]Provider{}
	for _, provider := range providers {
		byID[provider.ID] = provider
	}
	if byID["codex"].Status != StatusInstalled || byID["codex"].Version != "1.2.3" {
		t.Fatalf("codex detection = %#v", byID["codex"])
	}
	if byID["claude-code"].Status != StatusNeedsSetup || byID["claude-code"].Enabled {
		t.Fatalf("claude-code detection = %#v", byID["claude-code"])
	}
	if byID["opencode"].Status != StatusNotFound {
		t.Fatalf("opencode detection = %#v", byID["opencode"])
	}
	if byID["shell"].Status != StatusInstalled || !byID["shell"].Enabled {
		t.Fatalf("shell detection = %#v", byID["shell"])
	}
}

func TestBuildLaunchUsesWorkflowStepConfig(t *testing.T) {
	registry := DefaultAdapterRegistry()
	step := WorkflowStepConfig{
		ID:                   "coding",
		Provider:             "codex",
		Command:              "codex",
		WorkingDirectoryMode: "worktree",
		Environment:          map[string]string{"FROM_STEP": "1", "OVERRIDE": "step"},
	}
	launch, err := registry.BuildLaunch(context.Background(), step, LaunchRequest{
		TaskID:       "T-1",
		ProjectRoot:  "/repo",
		WorktreePath: "/repo/.thanos/worktrees/T-1",
		Args:         []string{"exec", "-"},
		Environment:  map[string]string{"OVERRIDE": "request"},
	})
	if err != nil {
		t.Fatalf("BuildLaunch: %v", err)
	}
	if launch.ProviderID != "codex" || launch.StepID != "coding" || launch.Command != "codex" {
		t.Fatalf("unexpected launch identity: %#v", launch)
	}
	if launch.CWD != "/repo/.thanos/worktrees/T-1" {
		t.Fatalf("cwd = %q", launch.CWD)
	}
	if !reflect.DeepEqual(launch.Args, []string{"exec", "-"}) {
		t.Fatalf("args = %#v", launch.Args)
	}
	if launch.Env["FROM_STEP"] != "1" || launch.Env["OVERRIDE"] != "request" {
		t.Fatalf("env merge = %#v", launch.Env)
	}
}

func TestBuildLaunchDoesNotExecuteAgent(t *testing.T) {
	registry := DefaultAdapterRegistry()
	launch, err := registry.BuildLaunch(context.Background(), WorkflowStepConfig{
		ID:                   "testing",
		Provider:             "shell",
		Command:              "npm test",
		WorkingDirectoryMode: "project",
	}, LaunchRequest{ProjectRoot: "/repo"})
	if err != nil {
		t.Fatalf("BuildLaunch: %v", err)
	}
	if launch.Command != "npm test" || launch.CWD != "/repo" {
		t.Fatalf("shell launch = %#v", launch)
	}
}

func TestBuildRestoreRequiresNativeSessionID(t *testing.T) {
	adapter, ok := DefaultAdapterRegistry().Get("codex")
	if !ok {
		t.Fatal("codex adapter missing")
	}
	if _, ok, err := adapter.BuildRestore(context.Background(), WorkflowStepConfig{ID: "coding", Command: "codex"}, RestoreRequest{}); err != nil || ok {
		t.Fatalf("empty restore = ok %v err %v", ok, err)
	}
	launch, ok, err := adapter.BuildRestore(context.Background(), WorkflowStepConfig{ID: "coding", Command: "codex"}, RestoreRequest{
		AgentNativeSessionID: "native-session",
		ProjectRoot:          "/repo",
	})
	if err != nil || !ok {
		t.Fatalf("restore = ok %v err %v", ok, err)
	}
	if launch.Command != "codex" {
		t.Fatalf("restore launch = %#v", launch)
	}
}
