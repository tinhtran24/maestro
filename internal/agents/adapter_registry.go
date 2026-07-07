package agents

import (
	"context"
	"fmt"
	"strings"
)

type AdapterRegistry struct {
	adapters map[string]AgentAdapter
	aliases  map[string]string
	order    []string
}

func NewAdapterRegistry(adapters ...AgentAdapter) *AdapterRegistry {
	registry := &AdapterRegistry{
		adapters: map[string]AgentAdapter{},
		aliases:  map[string]string{},
	}
	for _, adapter := range adapters {
		registry.Add(adapter)
	}
	return registry
}

func DefaultAdapterRegistry() *AdapterRegistry {
	registry := NewAdapterRegistry(
		NewProfiledCLIAdapter("claude-code", "Claude Code", "claude", "Install Claude Code and ensure `claude` is on PATH.", claudeCodeArgs),
		NewProfiledCLIAdapter("codex", "Codex", "codex", "Install Codex and ensure `codex` is on PATH.", codexArgs),
		NewProfiledCLIAdapter("gemini-cli", "Gemini CLI", "gemini", "Install Gemini CLI and ensure `gemini` is on PATH.", geminiArgs),
		NewProfiledCLIAdapter("opencode", "OpenCode", "opencode", "Install OpenCode and finish setup before assigning it.", opencodeArgs),
		NewProfiledCLIAdapter("cursor-agent", "Cursor Agent", "cursor-agent", "Install Cursor Agent and ensure `cursor-agent` is on PATH.", cursorAgentArgs),
		NewProfiledCLIAdapter("aider", "Aider", "aider", "Install Aider and ensure `aider` is on PATH.", aiderArgs),
		NewProfiledCLIAdapter("goose", "Goose", "goose", "Install Goose and ensure `goose` is on PATH.", gooseArgs),
		NewShellAdapter(),
	)
	registry.Alias("claude", "claude-code")
	registry.Alias("gemini", "gemini-cli")
	registry.Alias("cursor", "cursor-agent")
	registry.Alias("custom", "shell")
	return registry
}

func (r *AdapterRegistry) Add(adapter AgentAdapter) {
	if adapter == nil {
		return
	}
	id := normalizeProviderID(adapter.ID())
	if _, exists := r.adapters[id]; !exists {
		r.order = append(r.order, id)
	}
	r.adapters[id] = adapter
}

func (r *AdapterRegistry) Alias(alias, target string) {
	if r == nil {
		return
	}
	r.aliases[normalizeProviderID(alias)] = normalizeProviderID(target)
}

func (r *AdapterRegistry) Get(id string) (AgentAdapter, bool) {
	if r == nil {
		return nil, false
	}
	key := normalizeProviderID(id)
	if target, ok := r.aliases[key]; ok {
		key = target
	}
	adapter, ok := r.adapters[key]
	return adapter, ok
}

func (r *AdapterRegistry) All() []AgentAdapter {
	if r == nil {
		return nil
	}
	out := make([]AgentAdapter, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.adapters[id])
	}
	return out
}

func (r *AdapterRegistry) DetectAll(ctx context.Context, detection DetectionContext) ([]Provider, error) {
	var providers []Provider
	for _, adapter := range r.All() {
		provider, err := adapter.Detect(ctx, detection)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func (r *AdapterRegistry) BuildLaunch(ctx context.Context, step WorkflowStepConfig, req LaunchRequest) (LaunchCommand, error) {
	adapter, ok := r.Get(step.Provider)
	if !ok {
		adapter, ok = r.Get(step.Command)
	}
	if !ok {
		return LaunchCommand{}, fmt.Errorf("no agent adapter registered for provider %q", step.Provider)
	}
	return adapter.BuildLaunch(ctx, step, req)
}

func normalizeProviderID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

func claudeCodeArgs(step WorkflowStepConfig, req LaunchRequest) []string {
	args := []string{"--print"}
	args = appendModelArg(args, "--model", firstNonEmpty(req.Model, step.Model))
	if req.Prompt != "" {
		args = append(args, req.Prompt)
	}
	return args
}

func codexArgs(step WorkflowStepConfig, req LaunchRequest) []string {
	args := []string{"exec"}
	args = appendModelArg(args, "--model", firstNonEmpty(req.Model, step.Model))
	args = appendPermissions(args, step.Permissions)
	if req.Prompt != "" {
		args = append(args, req.Prompt)
	}
	return args
}

func geminiArgs(step WorkflowStepConfig, req LaunchRequest) []string {
	args := []string{"--prompt"}
	if req.Prompt != "" {
		args = append(args, req.Prompt)
	}
	return appendModelArg(args, "--model", firstNonEmpty(req.Model, step.Model))
}

func opencodeArgs(step WorkflowStepConfig, req LaunchRequest) []string {
	args := []string{"run"}
	args = appendModelArg(args, "--model", firstNonEmpty(req.Model, step.Model))
	if req.Prompt != "" {
		args = append(args, req.Prompt)
	}
	return args
}

func cursorAgentArgs(step WorkflowStepConfig, req LaunchRequest) []string {
	args := []string{}
	args = appendModelArg(args, "--model", firstNonEmpty(req.Model, step.Model))
	if req.Prompt != "" {
		args = append(args, req.Prompt)
	}
	return args
}

func aiderArgs(step WorkflowStepConfig, req LaunchRequest) []string {
	args := []string{"--message"}
	if req.Prompt != "" {
		args = append(args, req.Prompt)
	}
	return appendModelArg(args, "--model", firstNonEmpty(req.Model, step.Model))
}

func gooseArgs(step WorkflowStepConfig, req LaunchRequest) []string {
	args := []string{"run"}
	args = appendModelArg(args, "--model", firstNonEmpty(req.Model, step.Model))
	if req.Prompt != "" {
		args = append(args, "--text", req.Prompt)
	}
	return args
}

func appendModelArg(args []string, flag, model string) []string {
	model = strings.TrimSpace(model)
	if model == "" {
		return args
	}
	return append(args, flag, model)
}

func appendPermissions(args []string, permissions []string) []string {
	for _, permission := range permissions {
		permission = strings.TrimSpace(permission)
		if permission != "" {
			args = append(args, "--permission", permission)
		}
	}
	return args
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
