package agents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type AuthStatus string

const (
	AuthUnknown      AuthStatus = "unknown"
	AuthAuthorized   AuthStatus = "authorized"
	AuthUnauthorized AuthStatus = "unauthorized"
)

type PromptDeliveryMode string

const (
	PromptInArgs       PromptDeliveryMode = "in_args"
	PromptAfterStart   PromptDeliveryMode = "after_start"
	PromptFromWorkflow PromptDeliveryMode = "from_workflow"
)

type DetectionContext struct {
	LookPath func(command string) (string, error)
	Version  func(command string) (string, error)
	Auth     func(providerID string) (AuthStatus, error)
}

func DefaultDetectionContext() DetectionContext {
	return DetectionContext{
		LookPath: exec.LookPath,
		Version:  func(string) (string, error) { return "", nil },
		Auth:     func(string) (AuthStatus, error) { return AuthUnknown, nil },
	}
}

type LaunchRequest struct {
	TaskID       string
	ProjectRoot  string
	WorktreePath string
	Prompt       string
	Model        string
	Args         []string
	Environment  map[string]string
}

type RestoreRequest struct {
	TaskID               string
	ProjectRoot          string
	WorktreePath         string
	AgentNativeSessionID string
	TranscriptPath       string
	Args                 []string
	Environment          map[string]string
}

type LaunchCommand struct {
	ProviderID string            `json:"provider_id"`
	StepID     string            `json:"step_id"`
	Command    string            `json:"command"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	CWD        string            `json:"cwd,omitempty"`
}

type AgentAdapter interface {
	ID() string
	Name() string
	DefaultCommand() string
	Type() string
	Detect(ctx context.Context, detection DetectionContext) (Provider, error)
	BuildLaunch(ctx context.Context, step WorkflowStepConfig, req LaunchRequest) (LaunchCommand, error)
	BuildRestore(ctx context.Context, step WorkflowStepConfig, req RestoreRequest) (LaunchCommand, bool, error)
	PromptDelivery(ctx context.Context, step WorkflowStepConfig) PromptDeliveryMode
}

type CLIAdapter struct {
	id             string
	name           string
	command        string
	kind           string
	setupHint      string
	restoreArg     string
	promptDelivery PromptDeliveryMode
	argBuilder     func(WorkflowStepConfig, LaunchRequest) []string
}

func NewCLIAdapter(id, name, command, setupHint string) CLIAdapter {
	return CLIAdapter{
		id:             id,
		name:           name,
		command:        command,
		kind:           "cli",
		setupHint:      setupHint,
		restoreArg:     "",
		promptDelivery: PromptAfterStart,
	}
}

func NewProfiledCLIAdapter(id, name, command, setupHint string, builder func(WorkflowStepConfig, LaunchRequest) []string) CLIAdapter {
	adapter := NewCLIAdapter(id, name, command, setupHint)
	adapter.argBuilder = builder
	return adapter
}

func NewShellAdapter() CLIAdapter {
	return CLIAdapter{
		id:             "shell",
		name:           "Shell",
		command:        "sh",
		kind:           "shell",
		setupHint:      "Shell commands run through the configured workflow step command.",
		promptDelivery: PromptFromWorkflow,
	}
}

func (a CLIAdapter) ID() string {
	return a.id
}

func (a CLIAdapter) Name() string {
	return a.name
}

func (a CLIAdapter) DefaultCommand() string {
	return a.command
}

func (a CLIAdapter) Type() string {
	return a.kind
}

func (a CLIAdapter) Detect(_ context.Context, detection DetectionContext) (Provider, error) {
	if detection.LookPath == nil {
		detection = DefaultDetectionContext()
	}
	provider := Provider{
		ID:        a.ID(),
		Name:      a.Name(),
		Command:   a.DefaultCommand(),
		Type:      a.Type(),
		Status:    StatusNotFound,
		Enabled:   false,
		SetupHint: a.setupHint,
	}
	if a.Type() == "shell" {
		provider.Status = StatusInstalled
		provider.Enabled = true
		return provider, nil
	}
	path, err := detection.LookPath(a.DefaultCommand())
	if err != nil {
		return provider, nil
	}
	provider.DetectedPath = path
	provider.Status = StatusInstalled
	provider.Enabled = true
	if detection.Version != nil {
		version, err := detection.Version(a.DefaultCommand())
		if err == nil {
			provider.Version = version
		}
	}
	if detection.Auth != nil {
		auth, err := detection.Auth(a.ID())
		if err == nil && auth == AuthUnauthorized {
			provider.Status = StatusNeedsSetup
			provider.Enabled = false
			if provider.SetupHint == "" {
				provider.SetupHint = "Authenticate this CLI before assigning it to a workflow step."
			}
		}
	}
	return provider, nil
}

func (a CLIAdapter) BuildLaunch(_ context.Context, step WorkflowStepConfig, req LaunchRequest) (LaunchCommand, error) {
	command := strings.TrimSpace(step.Command)
	if command == "" {
		command = a.DefaultCommand()
	}
	if command == "" {
		return LaunchCommand{}, fmt.Errorf("workflow step %s has no command", step.ID)
	}
	args := append([]string(nil), req.Args...)
	if len(args) == 0 && a.argBuilder != nil {
		args = a.argBuilder(step, req)
	}
	return LaunchCommand{
		ProviderID: a.ID(),
		StepID:     step.ID,
		Command:    command,
		Args:       args,
		Env:        mergeEnv(step.Environment, req.Environment),
		CWD:        workingDirectory(step.WorkingDirectoryMode, req),
	}, nil
}

func (a CLIAdapter) BuildRestore(ctx context.Context, step WorkflowStepConfig, req RestoreRequest) (LaunchCommand, bool, error) {
	if req.AgentNativeSessionID == "" {
		return LaunchCommand{}, false, nil
	}
	launch, err := a.BuildLaunch(ctx, step, LaunchRequest{
		TaskID:       req.TaskID,
		ProjectRoot:  req.ProjectRoot,
		WorktreePath: req.WorktreePath,
		Args:         append([]string(nil), req.Args...),
		Environment:  req.Environment,
	})
	if err != nil {
		return LaunchCommand{}, false, err
	}
	if a.restoreArg != "" {
		launch.Args = append(launch.Args, a.restoreArg, req.AgentNativeSessionID)
	}
	return launch, true, nil
}

func (a CLIAdapter) PromptDelivery(_ context.Context, _ WorkflowStepConfig) PromptDeliveryMode {
	return a.promptDelivery
}

func mergeEnv(base, overlay map[string]string) map[string]string {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	out := make(map[string]string, len(base)+len(overlay))
	for key, value := range base {
		out[key] = value
	}
	for key, value := range overlay {
		out[key] = value
	}
	return out
}

func workingDirectory(mode string, req LaunchRequest) string {
	switch strings.TrimSpace(mode) {
	case "worktree":
		if req.WorktreePath != "" {
			return req.WorktreePath
		}
	case "project", "":
		return req.ProjectRoot
	}
	if req.WorktreePath != "" {
		return req.WorktreePath
	}
	return req.ProjectRoot
}
