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
		NewCLIAdapter("claude-code", "Claude Code", "claude", "Install Claude Code and ensure `claude` is on PATH."),
		NewCLIAdapter("codex", "Codex", "codex", "Install Codex and ensure `codex` is on PATH."),
		NewCLIAdapter("gemini-cli", "Gemini CLI", "gemini", "Install Gemini CLI and ensure `gemini` is on PATH."),
		NewCLIAdapter("opencode", "OpenCode", "opencode", "Install OpenCode and finish setup before assigning it."),
		NewShellAdapter(),
	)
	registry.Alias("claude", "claude-code")
	registry.Alias("gemini", "gemini-cli")
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
